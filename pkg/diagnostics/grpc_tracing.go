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
	"fmt"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	internalv1pb "github.com/bhojpur/api/pkg/core/v1/internals"
	runtimev1pb "github.com/bhojpur/api/pkg/core/v1/runtime"
	"github.com/bhojpur/application/pkg/config"
	diag_utils "github.com/bhojpur/application/pkg/diagnostics/utils"
)

const (
	grpcTraceContextKey = "grpc-trace-bin"
	GRPCProxyAppIDKey   = "app-app-id"
)

// GRPCTraceUnaryServerInterceptor sets the trace context or starts the trace client span based on request.
func GRPCTraceUnaryServerInterceptor(appID string, spec config.TracingSpec) grpc.UnaryServerInterceptor {
	sampler := diag_utils.TraceSampler(spec.SamplingRate)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var (
			span             *trace.Span
			spanKind         trace.StartOption
			prefixedMetadata map[string]string
			reqSpanAttr      map[string]string
		)
		sc, _ := SpanContextFromIncomingGRPCMetadata(ctx)
		// This middleware is shared by internal gRPC for service invocation and api
		// so that it needs to handle separately.
		if isInternalCalls(info.FullMethod) {
			// For pkg.core.v1.internals package, this generates ServerSpan.
			spanKind = trace.WithSpanKind(trace.SpanKindServer)
		} else {
			// For pkg.core.v1.runtime package, this generates ClientSpan.
			spanKind = trace.WithSpanKind(trace.SpanKindClient)
		}

		ctx, span = trace.StartSpanWithRemoteParent(ctx, info.FullMethod, sc, sampler, spanKind)

		isSampled := span.SpanContext().IsSampled()

		if isSampled {
			// users can add app- prefix if they want to see the header values in span attributes.
			prefixedMetadata = userDefinedMetadata(ctx)
			reqSpanAttr = spanAttributesMapFromGRPC(appID, req, info.FullMethod)
		}

		resp, err := handler(ctx, req)

		if isSampled {
			// Populates app- prefixed header first
			for key, value := range reqSpanAttr {
				prefixedMetadata[key] = value
			}
			AddAttributesToSpan(span, prefixedMetadata)

			// Correct the span name based on API.
			if sname, ok := reqSpanAttr[appAPISpanNameInternal]; ok {
				span.SetName(sname)
			}
		}

		// Add grpc-trace-bin header for all non-invocation api's
		if info.FullMethod != "/v1.runtime.Application/InvokeService" {
			traceContextBinary := propagation.Binary(span.SpanContext())
			grpc.SetHeader(ctx, metadata.Pairs(grpcTraceContextKey, string(traceContextBinary)))
		}

		UpdateSpanStatusFromGRPCError(span, err)
		span.End()

		return resp, err
	}
}

// GRPCTraceStreamServerInterceptor sets the trace context or starts the trace client span based on request.
func GRPCTraceStreamServerInterceptor(appID string, spec config.TracingSpec) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var span *trace.Span
		spanName := info.FullMethod

		ctx := ss.Context()
		md, _ := metadata.FromIncomingContext(ctx)

		vals := md.Get(GRPCProxyAppIDKey)
		if len(vals) == 0 {
			return errors.Errorf("cannot proxy request: missing %s metadata", GRPCProxyAppIDKey)
		}

		targetID := vals[0]
		wrapped := grpc_middleware.WrapServerStream(ss)
		sc, _ := SpanContextFromIncomingGRPCMetadata(ctx)
		sampler := diag_utils.TraceSampler(spec.SamplingRate)

		var spanKind trace.StartOption

		if appID == targetID {
			spanKind = trace.WithSpanKind(trace.SpanKindServer)
		} else {
			spanKind = trace.WithSpanKind(trace.SpanKindClient)
		}

		ctx, span = trace.StartSpanWithRemoteParent(ctx, spanName, sc, sampler, spanKind)
		wrapped.WrappedContext = ctx
		err := handler(srv, wrapped)

		addSpanMetadataAndUpdateStatus(ctx, span, info.FullMethod, appID, nil, true)

		UpdateSpanStatusFromGRPCError(span, err)
		span.End()

		return err
	}
}

func addSpanMetadataAndUpdateStatus(ctx context.Context, span *trace.Span, fullMethod, appID string, req interface{}, stream bool) {
	var prefixedMetadata map[string]string
	if span.SpanContext().TraceOptions.IsSampled() {
		// users can add app- prefix if they want to see the header values in span attributes.
		prefixedMetadata = userDefinedMetadata(ctx)
	}

	if span.SpanContext().TraceOptions.IsSampled() {
		// Populates app- prefixed header first
		AddAttributesToSpan(span, prefixedMetadata)

		spanAttr := map[string]string{}
		if !stream {
			spanAttr = spanAttributesMapFromGRPC(appID, req, fullMethod)
			AddAttributesToSpan(span, spanAttr)
		} else {
			spanAttr[appAPISpanNameInternal] = fullMethod
		}

		// Correct the span name based on API.
		if sname, ok := spanAttr[appAPISpanNameInternal]; ok {
			span.SetName(sname)
		}
	}
}

// userDefinedMetadata returns app- prefixed header from incoming metadata.
// Users can add app- prefixed headers that they want to see in span attributes.
func userDefinedMetadata(ctx context.Context) map[string]string {
	appMetadata := map[string]string{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return appMetadata
	}

	for k, v := range md {
		k = strings.ToLower(k)
		if strings.HasPrefix(k, appHeaderPrefix) && !strings.HasSuffix(k, appHeaderBinSuffix) {
			appMetadata[k] = v[0]
		}
	}

	return appMetadata
}

// UpdateSpanStatusFromGRPCError updates tracer span status based on error object.
func UpdateSpanStatusFromGRPCError(span *trace.Span, err error) {
	if span == nil || err == nil {
		return
	}

	s, ok := status.FromError(err)
	if ok {
		span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
	} else {
		span.SetStatus(trace.Status{Code: int32(codes.Internal), Message: err.Error()})
	}
}

// SpanContextFromIncomingGRPCMetadata returns the SpanContext stored in incoming metadata of context, or empty if there isn't one.
func SpanContextFromIncomingGRPCMetadata(ctx context.Context) (trace.SpanContext, bool) {
	var (
		sc trace.SpanContext
		md metadata.MD
		ok bool
	)
	if md, ok = metadata.FromIncomingContext(ctx); !ok {
		return sc, false
	}
	traceContext := md[grpcTraceContextKey]
	if len(traceContext) > 0 {
		sc, ok = propagation.FromBinary([]byte(traceContext[0]))
	} else {
		// add workaround to fallback on checking traceparent header
		// as grpc-trace-bin is not yet there in OpenTelemetry unlike OpenCensus , tracking issue https://github.com/open-telemetry/opentelemetry-specification/issues/639
		// and grpc-dotnet client adheres to OpenTelemetry Spec which only supports http based traceparent header in gRPC path
		// TODO : Remove this workaround fix once grpc-dotnet supports grpc-trace-bin header.
		traceContext = md[traceparentHeader]
		if len(traceContext) > 0 {
			sc, ok = SpanContextFromW3CString(traceContext[0])
			if ok && len(md[tracestateHeader]) > 0 {
				sc.Tracestate = TraceStateFromW3CString(md[tracestateHeader][0])
			}
		}
	}
	return sc, ok
}

// SpanContextToGRPCMetadata appends binary serialized SpanContext to the outgoing GRPC context.
func SpanContextToGRPCMetadata(ctx context.Context, spanContext trace.SpanContext) context.Context {
	traceContextBinary := propagation.Binary(spanContext)
	if len(traceContextBinary) == 0 {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, grpcTraceContextKey, string(traceContextBinary))
}

func isInternalCalls(method string) bool {
	return strings.HasPrefix(method, "/v1.internals.")
}

// spanAttributesMapFromGRPC builds the span trace attributes map for gRPC calls based on given parameters as per open-telemetry specs.
func spanAttributesMapFromGRPC(appID string, req interface{}, rpcMethod string) map[string]string {
	// RPC Span Attribute reference https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/rpc.md
	m := map[string]string{}

	var dbType string
	switch s := req.(type) {
	// Internal service invocation request
	case *internalv1pb.InternalInvokeRequest:
		m[gRPCServiceSpanAttributeKey] = appGRPCServiceInvocationService

		// Rename spanname
		if s.GetActor() == nil {
			m[appAPISpanNameInternal] = fmt.Sprintf("CallLocal/%s/%s", appID, s.Message.GetMethod())
			m[appAPIInvokeMethod] = s.Message.GetMethod()
		} else {
			m[appAPISpanNameInternal] = fmt.Sprintf("CallActor/%s/%s", s.GetActor().GetActorType(), s.Message.GetMethod())
			m[appAPIActorTypeID] = fmt.Sprintf("%s.%s", s.GetActor().GetActorType(), s.GetActor().GetActorId())
		}

	// Bhojpur Application APIs
	case *runtimev1pb.InvokeServiceRequest:
		m[gRPCServiceSpanAttributeKey] = appGRPCServiceInvocationService
		m[netPeerNameSpanAttributeKey] = s.GetId()
		m[appAPISpanNameInternal] = fmt.Sprintf("CallLocal/%s/%s", s.GetId(), s.Message.GetMethod())

	case *runtimev1pb.PublishEventRequest:
		m[gRPCServiceSpanAttributeKey] = appGRPCAppService
		m[messagingSystemSpanAttributeKey] = pubsubBuildingBlockType
		m[messagingDestinationSpanAttributeKey] = s.GetTopic()
		m[messagingDestinationKindSpanAttributeKey] = messagingDestinationTopicKind

	case *runtimev1pb.InvokeBindingRequest:
		dbType = bindingBuildingBlockType
		m[dbNameSpanAttributeKey] = s.GetName()

	case *runtimev1pb.GetStateRequest:
		dbType = stateBuildingBlockType
		m[dbNameSpanAttributeKey] = s.GetStoreName()

	case *runtimev1pb.SaveStateRequest:
		dbType = stateBuildingBlockType
		m[dbNameSpanAttributeKey] = s.GetStoreName()

	case *runtimev1pb.DeleteStateRequest:
		dbType = stateBuildingBlockType
		m[dbNameSpanAttributeKey] = s.GetStoreName()

	case *runtimev1pb.GetSecretRequest:
		dbType = secretBuildingBlockType
		m[dbNameSpanAttributeKey] = s.GetStoreName()
	}

	if _, ok := m[dbNameSpanAttributeKey]; ok {
		m[gRPCServiceSpanAttributeKey] = appGRPCAppService
		m[dbSystemSpanAttributeKey] = dbType
		m[dbStatementSpanAttributeKey] = rpcMethod
		m[dbConnectionStringSpanAttributeKey] = dbType
	}

	m[appAPIProtocolSpanAttributeKey] = appAPIGRPCSpanAttrValue
	m[appAPISpanAttributeKey] = rpcMethod

	return m
}
