package ca

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/bhojpur/service/pkg/utils/logger"

	"github.com/bhojpur/application/pkg/credentials"
	"github.com/bhojpur/application/pkg/sentry/certs"
	"github.com/bhojpur/application/pkg/sentry/config"
	"github.com/bhojpur/application/pkg/sentry/csr"
	"github.com/bhojpur/application/pkg/sentry/identity"
)

const (
	caOrg                      = "bhojpur.net/sentry"
	caCommonName               = "cluster.local"
	selfSignedRootCertLifetime = time.Hour * 8760
	certLoadTimeout            = time.Second * 30
	certDetectInterval         = time.Second * 1
)

var log = logger.NewLogger("app.sentry.ca")

// CertificateAuthority represents an interface for a compliant Certificate Authority.
// Responsibilities include loading trust anchors and issuer certs, providing safe access to the trust bundle,
// Validating and signing CSRs.
type CertificateAuthority interface {
	LoadOrStoreTrustBundle() error
	GetCACertBundle() TrustRootBundler
	SignCSR(csrPem []byte, subject string, identity *identity.Bundle, ttl time.Duration, isCA bool) (*SignedCertificate, error)
	ValidateCSR(csr *x509.CertificateRequest) error
}

func NewCertificateAuthority(config config.SentryConfig) (CertificateAuthority, error) {
	// Load future external CAs from components-contrib.
	switch config.CAStore {
	default:
		return &defaultCA{
			config:     config,
			issuerLock: &sync.RWMutex{},
		}, nil
	}
}

type defaultCA struct {
	bundle     *trustRootBundle
	config     config.SentryConfig
	issuerLock *sync.RWMutex
}

type SignedCertificate struct {
	Certificate *x509.Certificate
	CertPEM     []byte
}

// LoadOrStoreTrustBundle loads the root cert and issuer cert from the configured secret store.
// Validation is performed and a protected trust bundle is created holding the trust anchors
// and issuer credentials. If successful, a watcher is launched to keep track of the issuer expiration.
func (c *defaultCA) LoadOrStoreTrustBundle() error {
	bundle, err := c.validateAndBuildTrustBundle()
	if err != nil {
		return err
	}

	c.bundle = bundle
	return nil
}

// GetCACertBundle returns the Trust Root Bundle.
func (c *defaultCA) GetCACertBundle() TrustRootBundler {
	return c.bundle
}

// SignCSR signs a request with a PEM encoded CSR cert and duration.
// If isCA is set to true, a CA cert will be issued. If isCA is set to false, a workload
// Certificate will be issued instead.
func (c *defaultCA) SignCSR(csrPem []byte, subject string, identity *identity.Bundle, ttl time.Duration, isCA bool) (*SignedCertificate, error) {
	c.issuerLock.RLock()
	defer c.issuerLock.RUnlock()

	certLifetime := ttl
	if certLifetime.Seconds() < 0 {
		certLifetime = c.config.WorkloadCertTTL
	}

	certLifetime += c.config.AllowedClockSkew

	signingCert := c.bundle.issuerCreds.Certificate
	signingKey := c.bundle.issuerCreds.PrivateKey

	cert, err := certs.ParsePemCSR(csrPem)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing csr pem")
	}

	crtb, err := csr.GenerateCSRCertificate(cert, subject, identity, signingCert, cert.PublicKey, signingKey.Key, certLifetime, c.config.AllowedClockSkew, isCA)
	if err != nil {
		return nil, errors.Wrap(err, "error signing csr")
	}

	csrCert, err := x509.ParseCertificate(crtb)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing cert")
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  certs.Certificate,
		Bytes: crtb,
	})

	return &SignedCertificate{
		Certificate: csrCert,
		CertPEM:     certPem,
	}, nil
}

func (c *defaultCA) ValidateCSR(csr *x509.CertificateRequest) error {
	if csr.Subject.CommonName == "" {
		return errors.New("cannot validate request: missing common name")
	}
	return nil
}

func shouldCreateCerts(conf config.SentryConfig) bool {
	exists, err := certs.CredentialsExist(conf)
	if err != nil {
		log.Errorf("error checking if credentials exist: %s", err)
	}
	if exists {
		return false
	}

	if _, err = os.Stat(conf.RootCertPath); os.IsNotExist(err) {
		return true
	}
	fInfo, err := os.Stat(conf.IssuerCertPath)
	if os.IsNotExist(err) || fInfo.Size() == 0 {
		return true
	}
	return false
}

func detectCertificates(path string) error {
	t := time.NewTicker(certDetectInterval)
	timeout := time.After(certLoadTimeout)

	for {
		select {
		case <-timeout:
			return errors.New("timed out on detecting credentials on filesystem")
		case <-t.C:
			_, err := os.Stat(path)
			if err == nil {
				t.Stop()
				return nil
			}
		}
	}
}

func (c *defaultCA) validateAndBuildTrustBundle() (*trustRootBundle, error) {
	var (
		issuerCreds     *certs.Credentials
		rootCertBytes   []byte
		issuerCertBytes []byte
	)

	// certs exist on disk or getting created, load them when ready
	if !shouldCreateCerts(c.config) {
		err := detectCertificates(c.config.RootCertPath)
		if err != nil {
			return nil, err
		}

		certChain, err := credentials.LoadFromDisk(c.config.RootCertPath, c.config.IssuerCertPath, c.config.IssuerKeyPath)
		if err != nil {
			return nil, errors.Wrap(err, "error loading cert chain from disk")
		}

		issuerCreds, err = certs.PEMCredentialsFromFiles(certChain.Cert, certChain.Key)
		if err != nil {
			return nil, errors.Wrap(err, "error reading PEM credentials")
		}

		rootCertBytes = certChain.RootCA
		issuerCertBytes = certChain.Cert
	} else {
		// create self signed root and issuer certs
		log.Info("root and issuer certs not found: generating self signed CA")
		var err error
		issuerCreds, rootCertBytes, issuerCertBytes, err = c.generateRootAndIssuerCerts()
		if err != nil {
			return nil, errors.Wrap(err, "error generating trust root bundle")
		}

		log.Info("self signed certs generated and persisted successfully")
	}

	// load trust anchors
	trustAnchors, err := certs.CertPoolFromPEM(rootCertBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing cert pool for trust anchors")
	}

	return &trustRootBundle{
		issuerCreds:   issuerCreds,
		trustAnchors:  trustAnchors,
		trustDomain:   c.config.TrustDomain,
		rootCertPem:   rootCertBytes,
		issuerCertPem: issuerCertBytes,
	}, nil
}

func (c *defaultCA) generateRootAndIssuerCerts() (*certs.Credentials, []byte, []byte, error) {
	rootKey, err := certs.GenerateECPrivateKey()
	if err != nil {
		return nil, nil, nil, err
	}
	rootCsr, err := csr.GenerateRootCertCSR(caOrg, caCommonName, &rootKey.PublicKey, selfSignedRootCertLifetime, c.config.AllowedClockSkew)
	if err != nil {
		return nil, nil, nil, err
	}

	rootCertBytes, err := x509.CreateCertificate(rand.Reader, rootCsr, rootCsr, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, nil, nil, err
	}

	rootCertPem := pem.EncodeToMemory(&pem.Block{Type: certs.Certificate, Bytes: rootCertBytes})

	rootCert, err := x509.ParseCertificate(rootCertBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	issuerKey, err := certs.GenerateECPrivateKey()
	if err != nil {
		return nil, nil, nil, err
	}

	issuerCsr, err := csr.GenerateIssuerCertCSR(caCommonName, &issuerKey.PublicKey, selfSignedRootCertLifetime, c.config.AllowedClockSkew)
	if err != nil {
		return nil, nil, nil, err
	}

	issuerCertBytes, err := x509.CreateCertificate(rand.Reader, issuerCsr, rootCert, &issuerKey.PublicKey, rootKey)
	if err != nil {
		return nil, nil, nil, err
	}

	issuerCertPem := pem.EncodeToMemory(&pem.Block{Type: certs.Certificate, Bytes: issuerCertBytes})

	encodedKey, err := x509.MarshalECPrivateKey(issuerKey)
	if err != nil {
		return nil, nil, nil, err
	}
	issuerKeyPem := pem.EncodeToMemory(&pem.Block{Type: certs.ECPrivateKey, Bytes: encodedKey})

	issuerCert, err := x509.ParseCertificate(issuerCertBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	// store credentials so that next time sentry restarts it'll load normally
	err = certs.StoreCredentials(c.config, rootCertPem, issuerCertPem, issuerKeyPem)
	if err != nil {
		return nil, nil, nil, err
	}

	return &certs.Credentials{
		PrivateKey: &certs.PrivateKey{
			Type: certs.ECPrivateKey,
			Key:  issuerKey,
		},
		Certificate: issuerCert,
	}, rootCertPem, issuerCertPem, nil
}
