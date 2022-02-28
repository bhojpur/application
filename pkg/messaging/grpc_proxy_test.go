package messaging

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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/bhojpur/application/pkg/config"
	"github.com/bhojpur/application/pkg/diagnostics"
)

type sslEnabledConnection struct {
	sslEnabled bool
}

func (s *sslEnabledConnection) connectionSslFn(ctx context.Context, address, id string, namespace string, skipTLS, recreateIfExists, enableSSL bool, customOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	s.sslEnabled = enableSSL
	return grpc.Dial(id, grpc.WithInsecure())
}

func connectionFn(ctx context.Context, address, id string, namespace string, skipTLS, recreateIfExists, enableSSL bool, customOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(id, grpc.WithInsecure())
}

func TestNewProxy(t *testing.T) {
	p := NewProxy(connectionFn, "a", "a:123", 50005, nil, true)
	proxy := p.(*proxy)

	assert.Equal(t, "a", proxy.appID)
	assert.Equal(t, "a:123", proxy.localAppAddress)
	assert.Equal(t, 50005, proxy.remotePort)
	assert.NotNil(t, proxy.connectionFactory)
	assert.True(t, proxy.sslEnabled)
}

func TestSetRemoteAppFn(t *testing.T) {
	p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
	p.SetRemoteAppFn(func(s string) (remoteApp, error) {
		return remoteApp{
			id: "a",
		}, nil
	})

	proxy := p.(*proxy)
	app, err := proxy.remoteAppFn("a")

	assert.NoError(t, err)
	assert.Equal(t, "a", app.id)
}

func TestSetTelemetryFn(t *testing.T) {
	p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
	p.SetTelemetryFn(func(ctx context.Context) context.Context {
		return ctx
	})

	proxy := p.(*proxy)
	ctx := metadata.NewOutgoingContext(context.TODO(), metadata.MD{"a": []string{"b"}})
	ctx = proxy.telemetryFn(ctx)

	md, _ := metadata.FromOutgoingContext(ctx)
	assert.Equal(t, "b", md["a"][0])
}

func TestHandler(t *testing.T) {
	p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
	h := p.Handler()

	assert.NotNil(t, h)
}

func TestIntercept(t *testing.T) {
	t.Run("no app-id in metadata", func(t *testing.T) {
		p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
		p.SetTelemetryFn(func(ctx context.Context) context.Context {
			return ctx
		})

		p.SetRemoteAppFn(func(s string) (remoteApp, error) {
			return remoteApp{
				id: "a",
			}, nil
		})

		ctx := metadata.NewOutgoingContext(context.TODO(), metadata.MD{"a": []string{"b"}})
		proxy := p.(*proxy)
		_, conn, err := proxy.intercept(ctx, "/test")

		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("app-id exists in metadata", func(t *testing.T) {
		p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
		p.SetTelemetryFn(func(ctx context.Context) context.Context {
			return ctx
		})

		p.SetRemoteAppFn(func(s string) (remoteApp, error) {
			return remoteApp{
				id: "a",
			}, nil
		})

		ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{diagnostics.GRPCProxyAppIDKey: []string{"b"}})
		proxy := p.(*proxy)
		_, _, err := proxy.intercept(ctx, "/test")

		assert.NoError(t, err)
	})

	t.Run("proxy to the app", func(t *testing.T) {
		p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
		p.SetTelemetryFn(func(ctx context.Context) context.Context {
			return ctx
		})

		p.SetRemoteAppFn(func(s string) (remoteApp, error) {
			return remoteApp{
				id: "a",
			}, nil
		})

		ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{diagnostics.GRPCProxyAppIDKey: []string{"a"}})
		proxy := p.(*proxy)
		_, conn, err := proxy.intercept(ctx, "/test")

		assert.NoError(t, err)
		assert.NotNil(t, conn)
		assert.Equal(t, "a", conn.Target())
	})

	t.Run("proxy to a remote app", func(t *testing.T) {
		p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
		p.SetTelemetryFn(func(ctx context.Context) context.Context {
			ctx = metadata.AppendToOutgoingContext(ctx, "a", "b")
			return ctx
		})

		p.SetRemoteAppFn(func(s string) (remoteApp, error) {
			return remoteApp{
				id: "b",
			}, nil
		})

		ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{diagnostics.GRPCProxyAppIDKey: []string{"b"}})
		proxy := p.(*proxy)
		ctx, conn, err := proxy.intercept(ctx, "/test")

		assert.NoError(t, err)
		assert.NotNil(t, conn)
		assert.Equal(t, "b", conn.Target())

		md, _ := metadata.FromOutgoingContext(ctx)
		assert.Equal(t, "b", md["a"][0])
	})

	t.Run("access policies applied", func(t *testing.T) {
		acl := &config.AccessControlList{
			DefaultAction: "deny",
			TrustDomain:   "public",
		}

		p := NewProxy(connectionFn, "a", "a:123", 50005, acl, false)
		p.SetRemoteAppFn(func(s string) (remoteApp, error) {
			return remoteApp{
				id:      "a",
				address: "a:123",
			}, nil
		})
		p.SetTelemetryFn(func(ctx context.Context) context.Context {
			ctx = metadata.AppendToOutgoingContext(ctx, "a", "b")
			return ctx
		})

		ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{diagnostics.GRPCProxyAppIDKey: []string{"a"}})
		proxy := p.(*proxy)

		_, conn, err := proxy.intercept(ctx, "/test")

		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("SetRemoteAppFn never called", func(t *testing.T) {
		p := NewProxy(connectionFn, "a", "a:123", 50005, nil, false)
		p.SetTelemetryFn(func(ctx context.Context) context.Context {
			return ctx
		})

		ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{diagnostics.GRPCProxyAppIDKey: []string{"a"}})
		proxy := p.(*proxy)
		_, conn, err := proxy.intercept(ctx, "/test")

		assert.Error(t, err)
		assert.Nil(t, conn)
	})

	t.Run("ssl enabled", func(t *testing.T) {
		connFn := sslEnabledConnection{}

		p := NewProxy(connFn.connectionSslFn, "a", "a:123", 50005, nil, true)
		p.SetRemoteAppFn(func(s string) (remoteApp, error) {
			return remoteApp{
				id:      "a",
				address: "a:123",
			}, nil
		})
		p.SetTelemetryFn(func(ctx context.Context) context.Context {
			return ctx
		})

		ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{diagnostics.GRPCProxyAppIDKey: []string{"a"}})
		proxy := p.(*proxy)
		proxy.intercept(ctx, "/test")

		assert.True(t, connFn.sslEnabled)
	})
}
