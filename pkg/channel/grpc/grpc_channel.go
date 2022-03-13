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
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	internalv1pb "github.com/bhojpur/api/pkg/core/v1/internals"
	runtimev1pb "github.com/bhojpur/api/pkg/core/v1/runtime"
	"github.com/bhojpur/application/pkg/channel"
	"github.com/bhojpur/application/pkg/config"
	invokev1 "github.com/bhojpur/application/pkg/messaging/v1"
	auth "github.com/bhojpur/application/pkg/runtime/security"
)

// Channel is a concrete AppChannel implementation for interacting with gRPC based user code.
type Channel struct {
	client             *grpc.ClientConn
	baseAddress        string
	ch                 chan int
	tracingSpec        config.TracingSpec
	appMetadataToken   string
	maxRequestBodySize int
	readBufferSize     int
}

// CreateLocalChannel creates a gRPC connection with user code.
func CreateLocalChannel(port, maxConcurrency int, conn *grpc.ClientConn, spec config.TracingSpec, maxRequestBodySize int, readBufferSize int) *Channel {
	c := &Channel{
		client:             conn,
		baseAddress:        fmt.Sprintf("%s:%d", channel.DefaultChannelAddress, port),
		tracingSpec:        spec,
		appMetadataToken:   auth.GetAppToken(),
		maxRequestBodySize: maxRequestBodySize,
		readBufferSize:     readBufferSize,
	}
	if maxConcurrency > 0 {
		c.ch = make(chan int, maxConcurrency)
	}
	return c
}

// GetBaseAddress returns the application base address.
func (g *Channel) GetBaseAddress() string {
	return g.baseAddress
}

// GetAppConfig gets application config from user application.
func (g *Channel) GetAppConfig() (*config.ApplicationConfig, error) {
	return nil, nil
}

// InvokeMethod invokes user code via gRPC.
func (g *Channel) InvokeMethod(ctx context.Context, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {
	var rsp *invokev1.InvokeMethodResponse
	var err error

	switch req.APIVersion() {
	case internalv1pb.APIVersion_V1:
		rsp, err = g.invokeMethodV1(ctx, req)

	default:
		// Reject unsupported version
		rsp = nil
		err = status.Error(codes.Unimplemented, fmt.Sprintf("Unsupported spec version: %d", req.APIVersion()))
	}

	return rsp, err
}

// invokeMethodV1 calls user applications using appctl v1.
func (g *Channel) invokeMethodV1(ctx context.Context, req *invokev1.InvokeMethodRequest) (*invokev1.InvokeMethodResponse, error) {
	if g.ch != nil {
		g.ch <- 1
	}

	clientV1 := runtimev1pb.NewAppCallbackClient(g.client)
	grpcMetadata := invokev1.InternalMetadataToGrpcMetadata(ctx, req.Metadata(), true)

	if g.appMetadataToken != "" {
		grpcMetadata.Set(auth.APITokenHeader, g.appMetadataToken)
	}

	// Prepare gRPC Metadata
	ctx = metadata.NewOutgoingContext(context.Background(), grpcMetadata)

	var header, trailer metadata.MD

	var opts []grpc.CallOption
	opts = append(opts, grpc.Header(&header), grpc.Trailer(&trailer),
		grpc.MaxCallSendMsgSize(g.maxRequestBodySize*1024*1024), grpc.MaxCallRecvMsgSize(g.maxRequestBodySize*1024*1024))

	resp, err := clientV1.OnInvoke(ctx, req.Message(), opts...)

	if g.ch != nil {
		<-g.ch
	}

	var rsp *invokev1.InvokeMethodResponse
	if err != nil {
		// Convert status code
		respStatus := status.Convert(err)
		// Prepare response
		rsp = invokev1.NewInvokeMethodResponse(int32(respStatus.Code()), respStatus.Message(), respStatus.Proto().Details)
	} else {
		rsp = invokev1.NewInvokeMethodResponse(int32(codes.OK), "", nil)
	}

	rsp.WithHeaders(header).WithTrailers(trailer)

	return rsp.WithMessage(resp), nil
}
