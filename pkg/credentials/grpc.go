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
	"crypto/tls"
	"crypto/x509"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func GetServerOptions(certChain *CertChain) ([]grpc.ServerOption, error) {
	opts := []grpc.ServerOption{}
	if certChain == nil {
		return opts, nil
	}

	cp := x509.NewCertPool()
	cp.AppendCertsFromPEM(certChain.RootCA)

	cert, err := tls.X509KeyPair(certChain.Cert, certChain.Key)
	if err != nil {
		return opts, nil
	}

	// nolint:gosec
	config := &tls.Config{
		ClientCAs: cp,
		// Require cert verification
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
	}
	opts = append(opts, grpc.Creds(credentials.NewTLS(config)))

	return opts, nil
}

func GetClientOptions(certChain *CertChain, serverName string) ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{}
	if certChain != nil {
		cp := x509.NewCertPool()
		ok := cp.AppendCertsFromPEM(certChain.RootCA)
		if !ok {
			return nil, errors.New("failed to append PEM root cert to x509 CertPool")
		}
		config, err := TLSConfigFromCertAndKey(certChain.Cert, certChain.Key, serverName, cp)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tls config from cert and key")
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return opts, nil
}
