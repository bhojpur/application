package csr

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
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bhojpur/application/pkg/sentry/certs"
)

func TestGenerateCSRTemplate(t *testing.T) {
	t.Run("valid csr template", func(t *testing.T) {
		tmpl, err := genCSRTemplate("test-org")
		assert.Nil(t, err)
		assert.Equal(t, "test-org", tmpl.Subject.Organization[0])
	})
}

func TestGenerateBaseCertificate(t *testing.T) {
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cert, err := generateBaseCert(time.Second*5, time.Minute*15, pk)

	assert.NoError(t, err)
	assert.Equal(t, cert.PublicKey, pk)
}

func TestGenerateIssuerCertCSR(t *testing.T) {
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cert, err := GenerateIssuerCertCSR("name", pk, time.Second*5, time.Minute*15)

	assert.NoError(t, err)
	assert.Equal(t, "name", cert.DNSNames[0])
	assert.Equal(t, "name", cert.Subject.CommonName)
}

func TestGenerateRootCertCSR(t *testing.T) {
	pk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	cert, err := GenerateRootCertCSR("org", "name", pk, time.Second*5, time.Minute*15)

	assert.NoError(t, err)
	assert.Equal(t, "name", cert.Subject.CommonName)
	assert.Equal(t, "org", cert.Subject.Organization[0])
}

func TestCertSerialNumber(t *testing.T) {
	n, err := newSerialNumber()
	assert.NoError(t, err)
	assert.NotNil(t, n)
}

func TestGenerateCSR(t *testing.T) {
	c, b, err := GenerateCSR("org", false)
	assert.NoError(t, err)
	assert.True(t, len(b) > 0)
	assert.True(t, len(c) > 0)
}

func TestEncode(t *testing.T) {
	CertTests := []struct {
		org     string
		isPkcs8 bool
	}{
		{
			org:     "bhojpur.net",
			isPkcs8: false,
		},
		{
			org:     "bhojpur.net",
			isPkcs8: true,
		},
	}
	for _, test := range CertTests {
		key, err := certs.GenerateECPrivateKey()
		assert.Nil(t, err)
		template, err := genCSRTemplate(test.org)
		assert.Nil(t, err)
		csrBytes, err := x509.CreateCertificateRequest(rand.Reader, template, crypto.PrivateKey(key))
		assert.Nil(t, err)
		_, _, err = encode(true, csrBytes, key, test.isPkcs8)
		assert.Nil(t, err)
	}
}
