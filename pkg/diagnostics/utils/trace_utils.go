package utils

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
	"strconv"

	"github.com/valyala/fasthttp"
	"go.opencensus.io/trace"

	"github.com/bhojpur/service/pkg/utils/logger"
)

const (
	defaultSamplingRate = 1e-4

	// appFastHTTPContextKey is the context value of span in fasthttp.RequestCtx.
	appFastHTTPContextKey = "appSpanContextKey"
)

// StdoutExporter is an open census exporter that writes to stdout.
type StdoutExporter struct{}

var _ trace.Exporter = &StdoutExporter{}

var log = logger.NewLogger("app.runtime.trace")

const msg = "[%s] Trace: %s Span: %s/%s Time: [%s ->  %s] Annotations: %+v"

// ExportSpan implements the open census exporter interface.
func (e *StdoutExporter) ExportSpan(sd *trace.SpanData) {
	log.Infof(msg, sd.Name, sd.TraceID, sd.ParentSpanID, sd.SpanID, sd.StartTime, sd.EndTime, sd.Annotations)
}

// GetTraceSamplingRate parses the given rate and returns the parsed rate.
func GetTraceSamplingRate(rate string) float64 {
	f, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return defaultSamplingRate
	}
	return f
}

// TraceSampler returns Probability Sampler option.
func TraceSampler(samplingRate string) trace.StartOption {
	return trace.WithSampler(trace.ProbabilitySampler(GetTraceSamplingRate(samplingRate)))
}

// IsTracingEnabled parses the given rate and returns false if sampling rate is explicitly set 0.
func IsTracingEnabled(rate string) bool {
	return GetTraceSamplingRate(rate) != 0
}

// SpanFromContext returns the SpanContext stored in a context, or nil if there isn't one.
func SpanFromContext(ctx context.Context) *trace.Span {
	if reqCtx, ok := ctx.(*fasthttp.RequestCtx); ok {
		val := reqCtx.UserValue(appFastHTTPContextKey)
		if val == nil {
			return nil
		}
		return val.(*trace.Span)
	}

	return trace.FromContext(ctx)
}

// SpanToFastHTTPContext sets span into fasthttp.RequestCtx.
func SpanToFastHTTPContext(ctx *fasthttp.RequestCtx, span *trace.Span) {
	ctx.SetUserValue(appFastHTTPContextKey, span)
}
