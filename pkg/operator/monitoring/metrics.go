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
	appID = "app_id"
)

var (
	serviceCreatedTotal = stats.Int64(
		"operator/service_created_total",
		"The total number of Bhojpur Application services created.",
		stats.UnitDimensionless)
	serviceDeletedTotal = stats.Int64(
		"operator/service_deleted_total",
		"The total number of Bhojpur Application services deleted.",
		stats.UnitDimensionless)
	serviceUpdatedTotal = stats.Int64(
		"operator/service_updated_total",
		"The total number of Bhojpur Application services updated.",
		stats.UnitDimensionless)

	// appIDKey is a tag key for App ID.
	appIDKey = tag.MustNewKey(appID)
)

// RecordServiceCreatedCount records the number of Bhojpur Application service created.
func RecordServiceCreatedCount(appID string) {
	stats.RecordWithTags(context.Background(), diag_utils.WithTags(appIDKey, appID), serviceCreatedTotal.M(1))
}

// RecordServiceDeletedCount records the number of Bhojpur Application service deleted.
func RecordServiceDeletedCount(appID string) {
	stats.RecordWithTags(context.Background(), diag_utils.WithTags(appIDKey, appID), serviceDeletedTotal.M(1))
}

// RecordServiceUpdatedCount records the number of Bhojpur Application service updated.
func RecordServiceUpdatedCount(appID string) {
	stats.RecordWithTags(context.Background(), diag_utils.WithTags(appIDKey, appID), serviceUpdatedTotal.M(1))
}

// InitMetrics initialize the operator service metrics.
func InitMetrics() error {
	err := view.Register(
		diag_utils.NewMeasureView(serviceCreatedTotal, []tag.Key{appIDKey}, view.Count()),
		diag_utils.NewMeasureView(serviceDeletedTotal, []tag.Key{appIDKey}, view.Count()),
		diag_utils.NewMeasureView(serviceUpdatedTotal, []tag.Key{appIDKey}, view.Count()),
	)

	return err
}
