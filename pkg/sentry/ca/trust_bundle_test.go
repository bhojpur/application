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
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bhojpur/application/pkg/sentry/certs"
)

func getTestCert() *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization:  []string{"ORGANIZATION_NAME"},
			Country:       []string{"COUNTRY_CODE"},
			Province:      []string{"PROVINCE"},
			Locality:      []string{"CITY"},
			StreetAddress: []string{"ADDRESS"},
			PostalCode:    []string{"POSTAL_CODE"},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

func TestBundleIssuerExpiry(t *testing.T) {
	bundle := trustRootBundle{}
	issuerCert := getTestCert()
	bundle.issuerCreds = &certs.Credentials{
		Certificate: issuerCert,
	}

	assert.Equal(t, issuerCert.NotAfter.String(), bundle.GetIssuerCertExpiry().String())
}

func TestBundleIssuerCertMatch(t *testing.T) {
	bundle := trustRootBundle{}
	issuerCert := getTestCert()
	bundle.issuerCreds = &certs.Credentials{
		Certificate: issuerCert,
	}

	assert.Equal(t, issuerCert.Raw, bundle.GetIssuerCertPem())
}

func TestRootCertPEM(t *testing.T) {
	bundle := trustRootBundle{}
	cert := getTestCert()
	bundle.rootCertPem = cert.Raw

	assert.Equal(t, cert.Raw, bundle.GetRootCertPem())
}

func TestIssuerCertPEM(t *testing.T) {
	bundle := trustRootBundle{}
	cert := getTestCert()
	bundle.issuerCertPem = cert.Raw

	assert.Equal(t, cert.Raw, bundle.GetIssuerCertPem())
}

func TestTrustDomain(t *testing.T) {
	td := "td1"
	bundle := trustRootBundle{}
	bundle.trustDomain = td

	assert.Equal(t, td, bundle.GetTrustDomain())
}

func TestTrustAnchors(t *testing.T) {
	bundle := trustRootBundle{}
	pool := &x509.CertPool{}
	bundle.trustAnchors = pool

	assert.Equal(t, pool, bundle.GetTrustAnchors())
}
