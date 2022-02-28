package monitoring

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

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	diag_utils "github.com/bhojpur/application/pkg/diagnostics/utils"
)

var (
	runtimesTotal = stats.Int64(
		"placement/runtimes_total",
		"The total number of runtimes reported to placement service.",
		stats.UnitDimensionless)
	actorRuntimesTotal = stats.Int64(
		"placement/actor_runtimes_total",
		"The total number of actor runtimes reported to placement service.",
		stats.UnitDimensionless)

	noKeys = []tag.Key{}
)

// RecordRuntimesCount records the number of connected runtimes.
func RecordRuntimesCount(count int) {
	stats.Record(context.Background(), runtimesTotal.M(int64(count)))
}

// RecordActorRuntimesCount records the number of valid actor runtimes.
func RecordActorRuntimesCount(count int) {
	stats.Record(context.Background(), actorRuntimesTotal.M(int64(count)))
}

// InitMetrics initialize the placement service metrics.
func InitMetrics() error {
	err := view.Register(
		diag_utils.NewMeasureView(runtimesTotal, noKeys, view.LastValue()),
		diag_utils.NewMeasureView(actorRuntimesTotal, noKeys, view.LastValue()),
	)

	return err
}
