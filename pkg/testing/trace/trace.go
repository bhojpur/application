package trace

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
	"strconv"

	"go.opencensus.io/trace"

	"github.com/bhojpur/service/pkg/utils/logger"
)

// NewStringExporter returns a new string exporter instance.
//
// It is very useful in testing scenario where we want to validate trace propagation.
func NewStringExporter(buffer *string, logger logger.Logger) *Exporter {
	return &Exporter{
		Buffer: buffer,
		logger: logger,
	}
}

// Exporter is an OpenCensus string exporter.
type Exporter struct {
	Buffer *string
	logger logger.Logger
}

// ExportSpan exports span content to the buffer.
func (se *Exporter) ExportSpan(sd *trace.SpanData) {
	*se.Buffer = strconv.Itoa(int(sd.Status.Code))
}

// Register creates a new string exporter endpoint and reporter.
func (se *Exporter) Register(appID string) {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	trace.RegisterExporter(se)
}
