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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opencensus.io/trace"
)

func TestSpanFromContext(t *testing.T) {
	t.Run("fasthttp.RequestCtx, not nil span", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		SpanToFastHTTPContext(ctx, &trace.Span{})

		assert.NotNil(t, SpanFromContext(ctx))
	})

	t.Run("fasthttp.RequestCtx, nil span", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		SpanToFastHTTPContext(ctx, nil)

		assert.Nil(t, SpanFromContext(ctx))
	})

	t.Run("not nil span for context", func(t *testing.T) {
		ctx := context.Background()
		newCtx := trace.NewContext(ctx, &trace.Span{})

		assert.NotNil(t, SpanFromContext(newCtx))
	})

	t.Run("nil span for context", func(t *testing.T) {
		ctx := context.Background()
		newCtx := trace.NewContext(ctx, nil)

		assert.Nil(t, SpanFromContext(newCtx))
	})

	t.Run("nil", func(t *testing.T) {
		ctx := context.Background()

		assert.Nil(t, SpanFromContext(ctx))
	})
}
