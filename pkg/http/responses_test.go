package http

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

	"github.com/agrea/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestHeaders(t *testing.T) {
	t.Run("Respond with JSON", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{}}
		respond(ctx, withJSON(200, nil))

		assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
	})

	t.Run("Respond with JSON overrides custom content-type", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{}}
		ctx.Response.Header.SetContentType("custom")
		respond(ctx, withJSON(200, nil))

		assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
	})

	t.Run("Respond with ETag JSON", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{}}
		etagValue := "etagValue"
		respond(ctx, withJSON(200, nil), withEtag(ptr.String(etagValue)))

		assert.Equal(t, etagValue, string(ctx.Response.Header.Peek(etagHeader)))
	})

	t.Run("Respond with metadata and JSON", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{}}
		respond(ctx, withJSON(200, nil), withMetadata(map[string]string{"key": "value"}))

		assert.Equal(t, "value", string(ctx.Response.Header.Peek(metadataPrefix+"key")))
	})

	t.Run("Respond with custom content type", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{}}
		customContentType := "custom"
		ctx.Response.Header.SetContentType(customContentType)
		respond(ctx, with(200, nil))

		assert.Equal(t, customContentType, string(ctx.Response.Header.ContentType()))
	})

	t.Run("Respond with default content type", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{Request: fasthttp.Request{}}
		respond(ctx, with(200, nil))

		assert.Equal(t, "text/plain; charset=utf-8", string(ctx.Response.Header.ContentType()))
	})
}
