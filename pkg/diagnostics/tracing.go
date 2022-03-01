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
	"encoding/hex"
	"fmt"
	"strings"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/tracestate"

	"github.com/bhojpur/application/pkg/config"
	diag_utils "github.com/bhojpur/application/pkg/diagnostics/utils"

	// We currently don't depend on the Otel SDK since it has not GA'ed.
	// This package, however, only contains the conventions from the Otel Spec,
	// which we do depend on.
	// "go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/attribute"
	// "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	// "go.opentelemetry.io/otel/sdk/resource"
	// "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	appHeaderPrefix    = "app-"
	appHeaderBinSuffix = "-bin"

	// appInternalSpanAttrPrefix is the internal span attribution prefix.
	// Middleware will not populate it if the span key starts with this prefix.
	appInternalSpanAttrPrefix = "__app."
	// appAPISpanNameInternal is the internal attribution, but not populated
	// to span attribution.
	appAPISpanNameInternal = appInternalSpanAttrPrefix + "spanname"

	// span attribute keys
	// Reference trace semantics https://github.com/open-telemetry/opentelemetry-specification/tree/master/specification/trace/semantic_conventions
	//
	// The upstream constants may be used directly, but that would
	// proliferate the imports of go.opentelemetry.io/otel/... packages,
	// which we don't want to do widely before upstream goes GA.
	dbSystemSpanAttributeKey           = string(semconv.DBSystemKey)
	dbNameSpanAttributeKey             = string(semconv.DBNameKey)
	dbStatementSpanAttributeKey        = string(semconv.DBStatementKey)
	dbConnectionStringSpanAttributeKey = string(semconv.DBConnectionStringKey)

	messagingSystemSpanAttributeKey          = string(semconv.MessagingSystemKey)
	messagingDestinationSpanAttributeKey     = string(semconv.MessagingDestinationKey)
	messagingDestinationKindSpanAttributeKey = string(semconv.MessagingDestinationKindKey)

	gRPCServiceSpanAttributeKey = string(semconv.RPCServiceKey)
	netPeerNameSpanAttributeKey = string(semconv.NetPeerNameKey)

	appAPISpanAttributeKey           = "app.api"
	appAPIStatusCodeSpanAttributeKey = "app.status_code"
	appAPIProtocolSpanAttributeKey   = "app.protocol"
	appAPIInvokeMethod               = "app.invoke_method"
	appAPIActorTypeID                = "app.actor"

	appAPIHTTPSpanAttrValue = "http"
	appAPIGRPCSpanAttrValue = "grpc"

	stateBuildingBlockType   = "state"
	secretBuildingBlockType  = "secrets"
	bindingBuildingBlockType = "bindings"
	pubsubBuildingBlockType  = "pubsub"

	appGRPCServiceInvocationService = "ServiceInvocation"
	appGRPCAppService               = "App"
)

// Effectively const, but isn't a const from upstream.
var messagingDestinationTopicKind = semconv.MessagingDestinationKindTopic.Value.AsString()

// SpanContextToW3CString returns the SpanContext string representation.
func SpanContextToW3CString(sc trace.SpanContext) string {
	return fmt.Sprintf("%x-%x-%x-%x",
		[]byte{supportedVersion},
		sc.TraceID[:],
		sc.SpanID[:],
		[]byte{byte(sc.TraceOptions)})
}

// TraceStateToW3CString extracts the TraceState from given SpanContext and returns its string representation.
func TraceStateToW3CString(sc trace.SpanContext) string {
	pairs := make([]string, 0, len(sc.Tracestate.Entries()))
	if sc.Tracestate != nil {
		for _, entry := range sc.Tracestate.Entries() {
			pairs = append(pairs, strings.Join([]string{entry.Key, entry.Value}, "="))
		}
		h := strings.Join(pairs, ",")
		if h != "" && len(h) <= maxTracestateLen {
			return h
		}
	}
	return ""
}

// SpanContextFromW3CString extracts a span context from given string which got earlier from SpanContextToW3CString format.
func SpanContextFromW3CString(h string) (sc trace.SpanContext, ok bool) {
	if h == "" {
		return trace.SpanContext{}, false
	}
	sections := strings.Split(h, "-")
	if len(sections) < 4 {
		return trace.SpanContext{}, false
	}

	if len(sections[0]) != 2 {
		return trace.SpanContext{}, false
	}
	ver, err := hex.DecodeString(sections[0])
	if err != nil {
		return trace.SpanContext{}, false
	}
	version := int(ver[0])
	if version > maxVersion {
		return trace.SpanContext{}, false
	}

	if version == 0 && len(sections) != 4 {
		return trace.SpanContext{}, false
	}

	if len(sections[1]) != 32 {
		return trace.SpanContext{}, false
	}
	tid, err := hex.DecodeString(sections[1])
	if err != nil {
		return trace.SpanContext{}, false
	}
	copy(sc.TraceID[:], tid)

	if len(sections[2]) != 16 {
		return trace.SpanContext{}, false
	}
	sid, err := hex.DecodeString(sections[2])
	if err != nil {
		return trace.SpanContext{}, false
	}
	copy(sc.SpanID[:], sid)

	opts, err := hex.DecodeString(sections[3])
	if err != nil || len(opts) < 1 {
		return trace.SpanContext{}, false
	}
	sc.TraceOptions = trace.TraceOptions(opts[0])

	// Don't allow all zero trace or span ID.
	if sc.TraceID == [16]byte{} || sc.SpanID == [8]byte{} {
		return trace.SpanContext{}, false
	}

	return sc, true
}

// TraceStateFromW3CString extracts a span tracestate from given string which got earlier from TraceStateFromW3CString format.
func TraceStateFromW3CString(h string) *tracestate.Tracestate {
	if h == "" {
		return nil
	}

	entries := make([]tracestate.Entry, 0, len(h))
	pairs := strings.Split(h, ",")
	hdrLenWithoutOWS := len(pairs) - 1 // Number of commas
	for _, pair := range pairs {
		matches := trimOWSRegExp.FindStringSubmatch(pair)
		if matches == nil {
			return nil
		}
		pair = matches[1]
		hdrLenWithoutOWS += len(pair)
		if hdrLenWithoutOWS > maxTracestateLen {
			return nil
		}
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil
		}
		entries = append(entries, tracestate.Entry{Key: kv[0], Value: kv[1]})
	}
	ts, err := tracestate.New(nil, entries...)
	if err != nil {
		return nil
	}

	return ts
}

// AddAttributesToSpan adds the given attributes in the span.
func AddAttributesToSpan(span *trace.Span, attributes map[string]string) {
	if span == nil {
		return
	}

	var attrs []trace.Attribute
	for k, v := range attributes {
		// Skip if key is for internal use.
		if !strings.HasPrefix(k, appInternalSpanAttrPrefix) && v != "" {
			attrs = append(attrs, trace.StringAttribute(k, v))
		}
	}
	if len(attrs) > 0 {
		span.AddAttributes(attrs...)
	}
}

// ConstructInputBindingSpanAttributes creates span attributes for InputBindings.
func ConstructInputBindingSpanAttributes(bindingName, url string) map[string]string {
	return map[string]string{
		dbNameSpanAttributeKey:             bindingName,
		gRPCServiceSpanAttributeKey:        appGRPCAppService,
		dbSystemSpanAttributeKey:           bindingBuildingBlockType,
		dbConnectionStringSpanAttributeKey: url,
	}
}

// ConstructSubscriptionSpanAttributes creates span attributes for Pubsub subscription.
func ConstructSubscriptionSpanAttributes(topic string) map[string]string {
	return map[string]string{
		messagingSystemSpanAttributeKey:          pubsubBuildingBlockType,
		messagingDestinationSpanAttributeKey:     topic,
		messagingDestinationKindSpanAttributeKey: messagingDestinationTopicKind,
	}
}

// StartInternalCallbackSpan starts trace span for internal callback such as input bindings and pubsub subscription.
func StartInternalCallbackSpan(ctx context.Context, spanName string, parent trace.SpanContext, spec config.TracingSpec) (context.Context, *trace.Span) {
	traceEnabled := diag_utils.IsTracingEnabled(spec.SamplingRate)
	if !traceEnabled {
		return ctx, nil
	}

	sampler := diag_utils.TraceSampler(spec.SamplingRate)
	return trace.StartSpanWithRemoteParent(ctx, spanName, parent, sampler, trace.WithSpanKind(trace.SpanKindServer))
}
