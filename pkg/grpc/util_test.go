package grpc

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

	commonv1pb "github.com/bhojpur/api/pkg/core/v1/common"
)

func TestConsistency(t *testing.T) {
	t.Run("valid eventual", func(t *testing.T) {
		c := stateConsistencyToString(commonv1pb.StateOptions_CONSISTENCY_EVENTUAL)
		assert.Equal(t, "eventual", c)
	})

	t.Run("valid strong", func(t *testing.T) {
		c := stateConsistencyToString(commonv1pb.StateOptions_CONSISTENCY_STRONG)
		assert.Equal(t, "strong", c)
	})

	t.Run("empty when invalid", func(t *testing.T) {
		c := stateConsistencyToString(commonv1pb.StateOptions_CONSISTENCY_UNSPECIFIED)
		assert.Empty(t, c)
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("valid first write", func(t *testing.T) {
		c := stateConcurrencyToString(commonv1pb.StateOptions_CONCURRENCY_FIRST_WRITE)
		assert.Equal(t, "first-write", c)
	})

	t.Run("valid last write", func(t *testing.T) {
		c := stateConcurrencyToString(commonv1pb.StateOptions_CONCURRENCY_LAST_WRITE)
		assert.Equal(t, "last-write", c)
	})

	t.Run("empty when invalid", func(t *testing.T) {
		c := stateConcurrencyToString(commonv1pb.StateOptions_CONCURRENCY_UNSPECIFIED)
		assert.Empty(t, c)
	})
}
