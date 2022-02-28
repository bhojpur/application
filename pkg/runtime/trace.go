package runtime

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
	"go.opencensus.io/trace"
)

// traceExporterStore allows us to capture the trace exporter store registrations.
//
// This is needed because the OpenCensus library only expose global methods for
// exporter registration.
type traceExporterStore interface {
	// RegisterExporter registers a trace.Exporter.
	RegisterExporter(exporter trace.Exporter)
}

// openCensusExporterStore is an implementation of traceExporterStore
// that makes use of OpenCensus's library's global exporer stores (`trace`).
type openCensusExporterStore struct{}

// RegisterExporter implements traceExporterStore using OpenCensus's global registration.
func (s openCensusExporterStore) RegisterExporter(exporter trace.Exporter) {
	trace.RegisterExporter(exporter)
}

// fakeTraceExporterStore implements traceExporterStore by merely record the exporters
// and config that were registered/applied.
//
// This is only for use in unit tests.
type fakeTraceExporterStore struct {
	exporters []trace.Exporter
}

// RegisterExporter records the given trace.Exporter.
func (r *fakeTraceExporterStore) RegisterExporter(exporter trace.Exporter) {
	r.exporters = append(r.exporters, exporter)
}