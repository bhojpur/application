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

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// NewMeasureView creates opencensus View instance using stats.Measure.
func NewMeasureView(measure stats.Measure, keys []tag.Key, aggregation *view.Aggregation) *view.View {
	return &view.View{
		Name:        measure.Name(),
		Description: measure.Description(),
		Measure:     measure,
		TagKeys:     keys,
		Aggregation: aggregation,
	}
}

// WithTags converts tag key and value pairs to tag.Mutator array.
// WithTags(key1, value1, key2, value2) returns
// []tag.Mutator{tag.Upsert(key1, value1), tag.Upsert(key2, value2)}.
func WithTags(opts ...interface{}) []tag.Mutator {
	tagMutators := []tag.Mutator{}
	for i := 0; i < len(opts)-1; i += 2 {
		key, ok := opts[i].(tag.Key)
		if !ok {
			break
		}
		value, ok := opts[i+1].(string)
		if !ok {
			break
		}
		// skip if value is empty
		if value != "" {
			tagMutators = append(tagMutators, tag.Upsert(key, value))
		}
	}
	return tagMutators
}

// AddTagKeyToCtx assigns opencensus tag key value to context.
func AddTagKeyToCtx(ctx context.Context, key tag.Key, value string) context.Context {
	// return if value is not given
	if value == "" {
		return ctx
	}

	newCtx, err := tag.New(ctx, tag.Upsert(key, value))
	if err != nil {
		// return original if adding tagkey is failed.
		return ctx
	}

	return newCtx
}

// AddNewTagKey adds new tag keys to existing view.
func AddNewTagKey(views []*view.View, key *tag.Key) []*view.View {
	for _, v := range views {
		v.TagKeys = append(v.TagKeys, *key)
	}

	return views
}
