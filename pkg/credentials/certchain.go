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
	"os"
)

const (
	// RootCertFilename is the filename that holds the root certificate.
	RootCertFilename = "ca.crt"
	// IssuerCertFilename is the filename that holds the issuer certificate.
	IssuerCertFilename = "issuer.crt"
	// IssuerKeyFilename is the filename that holds the issuer key.
	IssuerKeyFilename = "issuer.key"
)

// CertChain holds the certificate trust chain PEM values.
type CertChain struct {
	RootCA []byte
	Cert   []byte
	Key    []byte
}

// LoadFromDisk retruns a CertChain from a given directory.
func LoadFromDisk(rootCertPath, issuerCertPath, issuerKeyPath string) (*CertChain, error) {
	rootCert, err := os.ReadFile(rootCertPath)
	if err != nil {
		return nil, err
	}
	cert, err := os.ReadFile(issuerCertPath)
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile(issuerKeyPath)
	if err != nil {
		return nil, err
	}
	return &CertChain{
		RootCA: rootCert,
		Cert:   cert,
		Key:    key,
	}, nil
}
