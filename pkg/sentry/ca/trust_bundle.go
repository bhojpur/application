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
	"time"

	"github.com/bhojpur/application/pkg/sentry/certs"
)

// TrustRootBundle represents the root certificate, issuer certificate and their
// Respective expiry dates.
type TrustRootBundler interface {
	GetIssuerCertPem() []byte
	GetRootCertPem() []byte
	GetIssuerCertExpiry() time.Time
	GetTrustAnchors() *x509.CertPool
	GetTrustDomain() string
}

type trustRootBundle struct {
	issuerCreds   *certs.Credentials
	trustAnchors  *x509.CertPool
	trustDomain   string
	rootCertPem   []byte
	issuerCertPem []byte
}

func (t *trustRootBundle) GetRootCertPem() []byte {
	return t.rootCertPem
}

func (t *trustRootBundle) GetIssuerCertPem() []byte {
	return t.issuerCertPem
}

func (t *trustRootBundle) GetIssuerCertExpiry() time.Time {
	return t.issuerCreds.Certificate.NotAfter
}

func (t *trustRootBundle) GetTrustAnchors() *x509.CertPool {
	return t.trustAnchors
}

func (t *trustRootBundle) GetTrustDomain() string {
	return t.trustDomain
}
