package grpc

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
	"context"
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/connectivity"

	"github.com/bhojpur/application/pkg/runtime/security"
	"github.com/bhojpur/application/pkg/utils"
)

type authenticatorMock struct{}

func (a *authenticatorMock) GetTrustAnchors() *x509.CertPool {
	return nil
}

func (a *authenticatorMock) GetCurrentSignedCert() *security.SignedCertificate {
	return nil
}

func (a *authenticatorMock) CreateSignedWorkloadCert(id, namespace, trustDomain string) (*security.SignedCertificate, error) {
	return nil, nil
}

func TestNewGRPCManager(t *testing.T) {
	t.Run("with self hosted", func(t *testing.T) {
		m := NewGRPCManager(utils.StandaloneMode)
		assert.NotNil(t, m)
		assert.Equal(t, utils.StandaloneMode, m.mode)
	})

	t.Run("with kubernetes", func(t *testing.T) {
		m := NewGRPCManager(utils.KubernetesMode)
		assert.NotNil(t, m)
		assert.Equal(t, utils.KubernetesMode, m.mode)
	})
}

func TestGetGRPCConnection(t *testing.T) {
	t.Run("Connection is closed", func(t *testing.T) {
		m := NewGRPCManager(utils.StandaloneMode)
		assert.NotNil(t, m)
		port := 55555
		sslEnabled := false
		ctx := context.TODO()
		conn, err := m.GetGRPCConnection(ctx, fmt.Sprintf("127.0.0.1:%v", port), "", "", true, true, sslEnabled)
		assert.NoError(t, err)
		conn2, err2 := m.GetGRPCConnection(ctx, fmt.Sprintf("127.0.0.1:%v", port), "", "", true, true, sslEnabled)
		assert.NoError(t, err2)
		assert.Equal(t, connectivity.Shutdown, conn.GetState())
		conn2.Close()
	})

	t.Run("Connection with SSL is created successfully", func(t *testing.T) {
		m := NewGRPCManager(utils.StandaloneMode)
		assert.NotNil(t, m)
		port := 55555
		sslEnabled := true
		ctx := context.TODO()
		_, err := m.GetGRPCConnection(ctx, fmt.Sprintf("127.0.0.1:%v", port), "", "", true, true, sslEnabled)
		assert.NoError(t, err)
	})
}

func TestSetAuthenticator(t *testing.T) {
	a := &authenticatorMock{}
	m := NewGRPCManager(utils.StandaloneMode)
	m.SetAuthenticator(a)

	assert.Equal(t, a, m.auth)
}
