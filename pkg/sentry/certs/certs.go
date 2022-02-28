package certs

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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/pkg/errors"
)

const (
	Certificate     = "CERTIFICATE"
	ECPrivateKey    = "EC PRIVATE KEY"
	RSAPrivateKey   = "RSA PRIVATE KEY"
	PKCS8PrivateKey = "PRIVATE KEY"
)

// PrivateKey wraps an EC or RSA private key.
type PrivateKey struct {
	Type string
	Key  interface{}
}

// Credentials holds a certificate, private key and trust chain.
type Credentials struct {
	PrivateKey  *PrivateKey
	Certificate *x509.Certificate
}

// DecodePEMKey takes a key PEM byte array and returns a PrivateKey that represents
// Either an RSA or EC private key.
func DecodePEMKey(key []byte) (*PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, errors.New("key is not PEM encoded")
	}
	switch block.Type {
	case ECPrivateKey:
		k, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &PrivateKey{Type: ECPrivateKey, Key: k}, nil
	case RSAPrivateKey:
		k, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &PrivateKey{Type: RSAPrivateKey, Key: k}, nil
	case PKCS8PrivateKey:
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return &PrivateKey{Type: PKCS8PrivateKey, Key: k}, nil
	default:
		return nil, errors.Errorf("unsupported block type %s", block.Type)
	}
}

// DecodePEMCertificates takes a PEM encoded x509 certificates byte array and returns
// A x509 certificate and the block byte array.
func DecodePEMCertificates(crtb []byte) ([]*x509.Certificate, error) {
	certs := []*x509.Certificate{}
	for len(crtb) > 0 {
		var err error
		var cert *x509.Certificate

		cert, crtb, err = decodeCertificatePEM(crtb)
		if err != nil {
			return nil, err
		}
		if cert != nil {
			// it's a cert, add to pool
			certs = append(certs, cert)
		}
	}
	return certs, nil
}

func decodeCertificatePEM(crtb []byte) (*x509.Certificate, []byte, error) {
	block, crtb := pem.Decode(crtb)
	if block == nil {
		return nil, crtb, errors.New("invalid PEM certificate")
	}
	if block.Type != Certificate {
		return nil, nil, nil
	}
	c, err := x509.ParseCertificate(block.Bytes)
	return c, crtb, err
}

// PEMCredentialsFromFiles takes a path for a key/cert pair and returns a validated Credentials wrapper with a trust chain.
func PEMCredentialsFromFiles(certPem, keyPem []byte) (*Credentials, error) {
	pk, err := DecodePEMKey(keyPem)
	if err != nil {
		return nil, err
	}

	crts, err := DecodePEMCertificates(certPem)
	if err != nil {
		return nil, err
	}

	if len(crts) == 0 {
		return nil, errors.New("no certificates found")
	}

	match := matchCertificateAndKey(pk, crts[0])
	if !match {
		return nil, errors.New("error validating credentials: public and private key pair do not match")
	}

	creds := &Credentials{
		PrivateKey:  pk,
		Certificate: crts[0],
	}

	return creds, nil
}

func matchCertificateAndKey(pk *PrivateKey, cert *x509.Certificate) bool {
	switch pk.Type {
	case ECPrivateKey:
		key := pk.Key.(*ecdsa.PrivateKey)
		pub, ok := cert.PublicKey.(*ecdsa.PublicKey)
		return ok && pub.X.Cmp(key.X) == 0 && pub.Y.Cmp(key.Y) == 0
	case RSAPrivateKey:
		key := pk.Key.(*rsa.PrivateKey)
		pub, ok := cert.PublicKey.(*rsa.PublicKey)
		return ok && pub.N.Cmp(key.N) == 0 && pub.E == key.E
	default:
		return false
	}
}

func certPoolFromCertificates(certs []*x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool
}

// CertPoolFromPEMString returns a CertPool from a PEM encoded certificates string.
func CertPoolFromPEM(certPem []byte) (*x509.CertPool, error) {
	certs, err := DecodePEMCertificates(certPem)
	if err != nil {
		return nil, err
	}
	if len(certs) == 0 {
		return nil, errors.New("no certificates found")
	}

	return certPoolFromCertificates(certs), nil
}

// ParsePemCSR constructs a x509 Certificate Request using the
// given PEM-encoded certificate signing request.
func ParsePemCSR(csrPem []byte) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode(csrPem)
	if block == nil {
		return nil, errors.New("certificate signing request is not properly encoded")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse X.509 certificate signing request")
	}
	return csr, nil
}

// GenerateECPrivateKey returns a new EC Private Key.
func GenerateECPrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}
