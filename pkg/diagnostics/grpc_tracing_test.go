package diagnostics

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
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	commonv1pb "github.com/bhojpur/application/pkg/api/v1/common"
	internalv1pb "github.com/bhojpur/application/pkg/api/v1/internals"
	runtimev1pb "github.com/bhojpur/application/pkg/api/v1/runtime"
	"github.com/bhojpur/application/pkg/config"
	diag_utils "github.com/bhojpur/application/pkg/diagnostics/utils"
)

func TestSpanAttributesMapFromGRPC(t *testing.T) {
	tests := []struct {
		rpcMethod                    string
		requestType                  string
		expectedServiceNameAttribute string
		expectedCustomAttribute      string
	}{
		{"/v1.internals.ServiceInvocation/CallLocal", "InternalInvokeRequest", "ServiceInvocation", "mymethod"},
		// InvokeService will be ServiceInvocation because this call will be treated as client call
		// of service invocation.
		{"/v1.runtime.Application/InvokeService", "InvokeServiceRequest", "ServiceInvocation", "mymethod"},
		{"/v1.runtime.Application/GetState", "GetStateRequest", "App", "mystore"},
		{"/v1.runtime.Application/SaveState", "SaveStateRequest", "App", "mystore"},
		{"/v1.runtime.Application/DeleteState", "DeleteStateRequest", "App", "mystore"},
		{"/v1.runtime.Application/GetSecret", "GetSecretRequest", "App", "mysecretstore"},
		{"/v1.runtime.Application/InvokeBinding", "InvokeBindingRequest", "App", "mybindings"},
		{"/v1.runtime.Application/PublishEvent", "PublishEventRequest", "App", "mytopic"},
	}
	var req interface{}
	for _, tt := range tests {
		t.Run(tt.rpcMethod, func(t *testing.T) {
			switch tt.requestType {
			case "InvokeServiceRequest":
				req = &runtimev1pb.InvokeServiceRequest{Message: &commonv1pb.InvokeRequest{Method: "mymethod"}}
			case "GetStateRequest":
				req = &runtimev1pb.GetStateRequest{StoreName: "mystore"}
			case "SaveStateRequest":
				req = &runtimev1pb.SaveStateRequest{StoreName: "mystore"}
			case "DeleteStateRequest":
				req = &runtimev1pb.DeleteStateRequest{StoreName: "mystore"}
			case "GetSecretRequest":
				req = &runtimev1pb.GetSecretRequest{StoreName: "mysecretstore"}
			case "InvokeBindingRequest":
				req = &runtimev1pb.InvokeBindingRequest{Name: "mybindings"}
			case "PublishEventRequest":
				req = &runtimev1pb.PublishEventRequest{Topic: "mytopic"}
			case "TopicEventRequest":
				req = &runtimev1pb.TopicEventRequest{Topic: "mytopic"}
			case "BindingEventRequest":
				req = &runtimev1pb.BindingEventRequest{Name: "mybindings"}
			case "InternalInvokeRequest":
				req = &internalv1pb.InternalInvokeRequest{Message: &commonv1pb.InvokeRequest{Method: "mymethod"}}
			}

			got := spanAttributesMapFromGRPC("fakeAppID", req, tt.rpcMethod)
			assert.Equal(t, tt.expectedServiceNameAttribute, got[gRPCServiceSpanAttributeKey], "servicename attribute should be equal")
		})
	}
}

func TestUserDefinedMetadata(t *testing.T) {
	md := metadata.MD{
		"app-userdefined-1": []string{"value1"},
		"app-userdefined-2": []string{"value2", "value3"},
		"no-attr":           []string{"value3"},
	}

	testCtx := metadata.NewIncomingContext(context.Background(), md)

	m := userDefinedMetadata(testCtx)

	assert.Equal(t, 2, len(m))
	assert.Equal(t, "value1", m["app-userdefined-1"])
	assert.Equal(t, "value2", m["app-userdefined-2"])
}

func TestSpanContextToGRPCMetadata(t *testing.T) {
	t.Run("empty span context", func(t *testing.T) {
		ctx := context.Background()
		newCtx := SpanContextToGRPCMetadata(ctx, trace.SpanContext{})

		assert.Equal(t, ctx, newCtx)
	})
}

func TestGRPCTraceUnaryServerInterceptor(t *testing.T) {
	rate := config.TracingSpec{SamplingRate: "1"}
	interceptor := GRPCTraceUnaryServerInterceptor("fakeAppID", rate)

	testTraceParent := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	testSpanContext, _ := SpanContextFromW3CString(testTraceParent)
	testTraceBinary := propagation.Binary(testSpanContext)
	ctx := context.Background()

	t.Run("grpc-trace-bin is given", func(t *testing.T) {
		ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("grpc-trace-bin", string(testTraceBinary)))
		fakeInfo := &grpc.UnaryServerInfo{
			FullMethod: "/v1.runtime.Application/GetState",
		}
		fakeReq := &runtimev1pb.GetStateRequest{
			StoreName: "statestore",
			Key:       "state",
		}

		var span *trace.Span
		assertHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			span = diag_utils.SpanFromContext(ctx)
			return nil, errors.New("fake error")
		}

		interceptor(ctx, fakeReq, fakeInfo, assertHandler)

		sc := span.SpanContext()
		assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", fmt.Sprintf("%x", sc.TraceID[:]))
		assert.NotEqual(t, "00f067aa0ba902b7", fmt.Sprintf("%x", sc.SpanID[:]))
	})

	t.Run("grpc-trace-bin is not given", func(t *testing.T) {
		fakeInfo := &grpc.UnaryServerInfo{
			FullMethod: "/v1.runtime.Application/GetState",
		}
		fakeReq := &runtimev1pb.GetStateRequest{
			StoreName: "statestore",
			Key:       "state",
		}

		var span *trace.Span
		assertHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			span = diag_utils.SpanFromContext(ctx)
			return nil, errors.New("fake error")
		}

		interceptor(ctx, fakeReq, fakeInfo, assertHandler)

		sc := span.SpanContext()
		assert.NotEmpty(t, fmt.Sprintf("%x", sc.TraceID[:]))
		assert.NotEmpty(t, fmt.Sprintf("%x", sc.SpanID[:]))
	})

	t.Run("InvokeService call", func(t *testing.T) {
		fakeInfo := &grpc.UnaryServerInfo{
			FullMethod: "/v1.runtime.Application/InvokeService",
		}
		fakeReq := &runtimev1pb.InvokeServiceRequest{
			Id:      "targetID",
			Message: &commonv1pb.InvokeRequest{Method: "method1"},
		}

		var span *trace.Span
		assertHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			span = diag_utils.SpanFromContext(ctx)
			return nil, errors.New("fake error")
		}

		interceptor(ctx, fakeReq, fakeInfo, assertHandler)

		sc := span.SpanContext()
		assert.True(t, strings.Contains(span.String(), "CallLocal/targetID/method1"))
		assert.NotEmpty(t, fmt.Sprintf("%x", sc.TraceID[:]))
		assert.NotEmpty(t, fmt.Sprintf("%x", sc.SpanID[:]))
	})
}

func TestSpanContextSerialization(t *testing.T) {
	wantSc := trace.SpanContext{
		TraceID:      trace.TraceID{75, 249, 47, 53, 119, 179, 77, 166, 163, 206, 146, 157, 14, 14, 71, 54},
		SpanID:       trace.SpanID{0, 240, 103, 170, 11, 169, 2, 183},
		TraceOptions: trace.TraceOptions(1),
	}

	passedOverWire := string(propagation.Binary(wantSc))
	storedInApp := base64.StdEncoding.EncodeToString([]byte(passedOverWire))
	decoded, _ := base64.StdEncoding.DecodeString(storedInApp)
	gotSc, _ := propagation.FromBinary(decoded)
	assert.Equal(t, wantSc, gotSc)
}
