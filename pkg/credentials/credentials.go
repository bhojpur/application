package credentials

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
	"path/filepath"
)

// TLSCredentials holds paths for credentials.
type TLSCredentials struct {
	credentialsPath string
}

// NewTLSCredentials returns a new TLSCredentials.
func NewTLSCredentials(path string) TLSCredentials {
	return TLSCredentials{
		credentialsPath: path,
	}
}

// Path returns the directory holding the TLS credentials.
func (t *TLSCredentials) Path() string {
	return t.credentialsPath
}

// RootCertPath returns the file path for the root cert.
func (t *TLSCredentials) RootCertPath() string {
	return filepath.Join(t.credentialsPath, RootCertFilename)
}

// CertPath returns the file path for the cert.
func (t *TLSCredentials) CertPath() string {
	return filepath.Join(t.credentialsPath, IssuerCertFilename)
}

// KeyPath returns the file path for the cert key.
func (t *TLSCredentials) KeyPath() string {
	return filepath.Join(t.credentialsPath, IssuerKeyFilename)
}
