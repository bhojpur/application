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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opencensus.io/tag"
)

func TestWithTags(t *testing.T) {
	t.Run("one tag", func(t *testing.T) {
		appKey := tag.MustNewKey("app_id")
		mutators := WithTags(appKey, "test")
		assert.Equal(t, 1, len(mutators))
	})

	t.Run("two tags", func(t *testing.T) {
		appKey := tag.MustNewKey("app_id")
		operationKey := tag.MustNewKey("operation")
		mutators := WithTags(appKey, "test", operationKey, "op")
		assert.Equal(t, 2, len(mutators))
	})

	t.Run("three tags", func(t *testing.T) {
		appKey := tag.MustNewKey("app_id")
		operationKey := tag.MustNewKey("operation")
		methodKey := tag.MustNewKey("method")
		mutators := WithTags(appKey, "test", operationKey, "op", methodKey, "method")
		assert.Equal(t, 3, len(mutators))
	})

	t.Run("two tags with wrong value type", func(t *testing.T) {
		appKey := tag.MustNewKey("app_id")
		operationKey := tag.MustNewKey("operation")
		mutators := WithTags(appKey, "test", operationKey, 1)
		assert.Equal(t, 1, len(mutators))
	})

	t.Run("skip empty value key", func(t *testing.T) {
		appKey := tag.MustNewKey("app_id")
		operationKey := tag.MustNewKey("operation")
		methodKey := tag.MustNewKey("method")
		mutators := WithTags(appKey, "", operationKey, "op", methodKey, "method")
		assert.Equal(t, 2, len(mutators))
	})
}
