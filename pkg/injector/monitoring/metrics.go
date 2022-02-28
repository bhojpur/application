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

const (
	appID        = "app_id"
	failedReason = "reason"
)

var (
	sidecarInjectionRequestsTotal = stats.Int64(
		"injector/sidecar_injection/requests_total",
		"The total number of sidecar injection requests.",
		stats.UnitDimensionless)
	succeededSidecarInjectedTotal = stats.Int64(
		"injector/sidecar_injection/succeeded_total",
		"The total number of successful sidecar injections.",
		stats.UnitDimensionless)
	failedSidecarInjectedTotal = stats.Int64(
		"injector/sidecar_injection/failed_total",
		"The total number of failed sidecar injections.",
		stats.UnitDimensionless)

	noKeys = []tag.Key{}

	// appIDKey is a tag key for App ID.
	appIDKey = tag.MustNewKey(appID)

	// failedReasonKey is a tag key for failed reason.
	failedReasonKey = tag.MustNewKey(failedReason)
)

// RecordSidecarInjectionRequestsCount records the total number of sidecar injection requests.
func RecordSidecarInjectionRequestsCount() {
	stats.Record(context.Background(), sidecarInjectionRequestsTotal.M(1))
}

// RecordSuccessfulSidecarInjectionCount records the number of successful sidecar injections.
func RecordSuccessfulSidecarInjectionCount(appID string) {
	stats.RecordWithTags(context.Background(), diag_utils.WithTags(appIDKey, appID), succeededSidecarInjectedTotal.M(1))
}

// RecordFailedSidecarInjectionCount records the number of failed sidecar injections.
func RecordFailedSidecarInjectionCount(appID, reason string) {
	stats.RecordWithTags(context.Background(), diag_utils.WithTags(appIDKey, appID, failedReasonKey, reason), failedSidecarInjectedTotal.M(1))
}

// InitMetrics initialize the injector service metrics.
func InitMetrics() error {
	err := view.Register(
		diag_utils.NewMeasureView(sidecarInjectionRequestsTotal, noKeys, view.Count()),
		diag_utils.NewMeasureView(succeededSidecarInjectedTotal, []tag.Key{appIDKey}, view.Count()),
		diag_utils.NewMeasureView(failedSidecarInjectedTotal, []tag.Key{appIDKey, failedReasonKey}, view.Count()),
	)

	return err
}
