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

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpc_proxy "github.com/bhojpur/application/pkg/grpc/proxy"
	codec "github.com/bhojpur/application/pkg/grpc/proxy/codec"

	"github.com/bhojpur/application/pkg/acl"
	"github.com/bhojpur/application/pkg/api/v1/common"
	"github.com/bhojpur/application/pkg/config"
	"github.com/bhojpur/application/pkg/diagnostics"
)

// Proxy is the interface for a gRPC transparent proxy.
type Proxy interface {
	Handler() grpc.StreamHandler
	SetRemoteAppFn(func(string) (remoteApp, error))
	SetTelemetryFn(func(context.Context) context.Context)
}

type proxy struct {
	appID             string
	connectionFactory messageClientConnection
	remoteAppFn       func(appID string) (remoteApp, error)
	remotePort        int
	telemetryFn       func(context.Context) context.Context
	localAppAddress   string
	acl               *config.AccessControlList
	sslEnabled        bool
}

// NewProxy returns a new proxy.
func NewProxy(connectionFactory messageClientConnection, appID string, localAppAddress string, remoteAppPort int, acl *config.AccessControlList, sslEnabled bool) Proxy {
	return &proxy{
		appID:             appID,
		connectionFactory: connectionFactory,
		localAppAddress:   localAppAddress,
		remotePort:        remoteAppPort,
		acl:               acl,
		sslEnabled:        sslEnabled,
	}
}

// Handler returns a Stream Handler for handling requests that arrive for services that are not recognized by the server.
func (p *proxy) Handler() grpc.StreamHandler {
	return grpc_proxy.TransparentHandler(p.intercept)
}

func (p *proxy) intercept(ctx context.Context, fullName string) (context.Context, *grpc.ClientConn, error) {
	md, _ := metadata.FromIncomingContext(ctx)

	v := md.Get(diagnostics.GRPCProxyAppIDKey)
	if len(v) == 0 {
		return ctx, nil, errors.Errorf("failed to proxy request: required metadata %s not found", diagnostics.GRPCProxyAppIDKey)
	}

	outCtx := metadata.NewOutgoingContext(ctx, md.Copy())
	appID := v[0]

	if p.remoteAppFn == nil {
		return ctx, nil, errors.Errorf("failed to proxy request: proxy not initialized. Bhojpur Application runtime startup may be incomplete.")
	}

	target, err := p.remoteAppFn(appID)
	if err != nil {
		return ctx, nil, err
	}

	if target.id == p.appID {
		// proxy locally to the app
		if p.acl != nil {
			ok, authError := acl.ApplyAccessControlPolicies(ctx, fullName, common.HTTPExtension_NONE, config.GRPCProtocol, p.acl)
			if !ok {
				return ctx, nil, status.Errorf(codes.PermissionDenied, authError)
			}
		}

		conn, cErr := p.connectionFactory(outCtx, p.localAppAddress, p.appID, "", true, false, p.sslEnabled, grpc.WithDefaultCallOptions(grpc.CallContentSubtype((&codec.Proxy{}).Name())))
		return outCtx, conn, cErr
	}

	// proxy to a remote Bhojpur Application runtime
	// connection is recreated because its certification may have already been expired
	conn, cErr := p.connectionFactory(outCtx, target.address, target.id, target.namespace, false, true, false, grpc.WithDefaultCallOptions(grpc.CallContentSubtype((&codec.Proxy{}).Name())))
	outCtx = p.telemetryFn(outCtx)

	return outCtx, conn, cErr
}

// SetRemoteAppFn sets a function that helps the proxy resolve an app ID to an actual address.
func (p *proxy) SetRemoteAppFn(remoteAppFn func(appID string) (remoteApp, error)) {
	p.remoteAppFn = remoteAppFn
}

// SetTelemetryFn sets a function that enriches the context with telemetry.
func (p *proxy) SetTelemetryFn(spanFn func(context.Context) context.Context) {
	p.telemetryFn = spanFn
}
