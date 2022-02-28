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
	"crypto/x509"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockGenCSR(id string) ([]byte, []byte, error) {
	return []byte{1}, []byte{2}, nil
}

func getTestAuthenticator() Authenticator {
	return newAuthenticator("test", x509.NewCertPool(), nil, nil, mockGenCSR)
}

func TestGetTrustAuthAnchors(t *testing.T) {
	a := getTestAuthenticator()
	ta := a.GetTrustAnchors()
	assert.NotNil(t, ta)
}

func TestGetCurrentSignedCert(t *testing.T) {
	a := getTestAuthenticator()
	a.(*authenticator).currentSignedCert = &SignedCertificate{}
	c := a.GetCurrentSignedCert()
	assert.NotNil(t, c)
}

func TestGetSentryIdentifier(t *testing.T) {
	t.Run("with identity in env", func(t *testing.T) {
		envID := "cluster.local"
		os.Setenv("SENTRY_LOCAL_IDENTITY", envID)
		defer os.Unsetenv("SENTRY_LOCAL_IDENTITY")

		id := getSentryIdentifier("app1")
		assert.Equal(t, envID, id)
	})

	t.Run("without identity in env", func(t *testing.T) {
		id := getSentryIdentifier("app1")
		assert.Equal(t, "app1", id)
	})
}
