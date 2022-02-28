package security

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
	"crypto/x509/pkix"
	"encoding/pem"
	"os"

	"github.com/pkg/errors"

	"github.com/bhojpur/service/pkg/utils/logger"

	"github.com/bhojpur/application/pkg/credentials"
	diag "github.com/bhojpur/application/pkg/diagnostics"
	"github.com/bhojpur/application/pkg/sentry/certs"
)

const (
	ecPKType = "EC PRIVATE KEY"
)

var log = logger.NewLogger("app.runtime.security")

func CertPool(certPem []byte) (*x509.CertPool, error) {
	cp := x509.NewCertPool()
	ok := cp.AppendCertsFromPEM(certPem)
	if !ok {
		return nil, errors.New("failed to append PEM root cert to x509 CertPool")
	}
	return cp, nil
}

func GetCertChain() (*credentials.CertChain, error) {
	trustAnchors := os.Getenv(certs.TrustAnchorsEnvVar)
	if trustAnchors == "" {
		return nil, errors.Errorf("couldn't find trust anchors in environment variable %s", certs.TrustAnchorsEnvVar)
	}
	cert := os.Getenv(certs.CertChainEnvVar)
	if cert == "" {
		return nil, errors.Errorf("couldn't find cert chain in environment variable %s", certs.CertChainEnvVar)
	}
	key := os.Getenv(certs.CertKeyEnvVar)
	if cert == "" {
		return nil, errors.Errorf("couldn't find cert key in environment variable %s", certs.CertKeyEnvVar)
	}
	return &credentials.CertChain{
		RootCA: []byte(trustAnchors),
		Cert:   []byte(cert),
		Key:    []byte(key),
	}, nil
}

// GetSidecarAuthenticator returns a new authenticator with the extracted trust anchors.
func GetSidecarAuthenticator(sentryAddress string, certChain *credentials.CertChain) (Authenticator, error) {
	trustAnchors, err := CertPool(certChain.RootCA)
	if err != nil {
		return nil, err
	}
	log.Info("trust anchors and cert chain extracted successfully")

	return newAuthenticator(sentryAddress, trustAnchors, certChain.Cert, certChain.Key, generateCSRAndPrivateKey), nil
}

func generateCSRAndPrivateKey(id string) ([]byte, []byte, error) {
	if id == "" {
		return nil, nil, errors.New("id must not be empty")
	}

	key, err := certs.GenerateECPrivateKey()
	if err != nil {
		diag.DefaultMonitoring.MTLSInitFailed("prikeygen")
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	encodedKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		diag.DefaultMonitoring.MTLSInitFailed("prikeyenc")
		return nil, nil, err
	}
	keyPem := pem.EncodeToMemory(&pem.Block{Type: ecPKType, Bytes: encodedKey})

	csr := x509.CertificateRequest{
		Subject:  pkix.Name{CommonName: id},
		DNSNames: []string{id},
	}
	csrb, err := x509.CreateCertificateRequest(rand.Reader, &csr, key)
	if err != nil {
		diag.DefaultMonitoring.MTLSInitFailed("csr")
		return nil, nil, errors.Wrap(err, "failed to create sidecar csr")
	}
	return csrb, keyPem, nil
}
