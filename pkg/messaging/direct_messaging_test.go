package messaging

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
	"github.com/valyala/fasthttp"

	invokev1 "github.com/bhojpur/application/pkg/messaging/v1"
)

func newDirectMessaging() *directMessaging {
	return &directMessaging{}
}

func TestDestinationHeaders(t *testing.T) {
	t.Run("destination header present", func(t *testing.T) {
		appID := "test1"
		req := invokev1.NewInvokeMethodRequest("GET")
		req.WithMetadata(map[string][]string{})

		dm := newDirectMessaging()
		dm.addDestinationAppIDHeaderToMetadata(appID, req)
		md := req.Metadata()[invokev1.DestinationIDHeader]
		assert.Equal(t, appID, md.Values[0])
	})
}

func TestForwardedHeaders(t *testing.T) {
	t.Run("forwarded headers present", func(t *testing.T) {
		req := invokev1.NewInvokeMethodRequest("GET")
		req.WithMetadata(map[string][]string{})

		dm := newDirectMessaging()
		dm.hostAddress = "1"
		dm.hostName = "2"

		dm.addForwardedHeadersToMetadata(req)

		md := req.Metadata()[fasthttp.HeaderXForwardedFor]
		assert.Equal(t, "1", md.Values[0])

		md = req.Metadata()[fasthttp.HeaderXForwardedHost]
		assert.Equal(t, "2", md.Values[0])

		md = req.Metadata()[fasthttp.HeaderForwarded]
		assert.Equal(t, "for=1;by=1;host=2", md.Values[0])
	})

	t.Run("forwarded headers get appended", func(t *testing.T) {
		req := invokev1.NewInvokeMethodRequest("GET")
		req.WithMetadata(map[string][]string{
			fasthttp.HeaderXForwardedFor:  {"originalXForwardedFor"},
			fasthttp.HeaderXForwardedHost: {"originalXForwardedHost"},
			fasthttp.HeaderForwarded:      {"originalForwarded"},
		})

		dm := newDirectMessaging()
		dm.hostAddress = "1"
		dm.hostName = "2"

		dm.addForwardedHeadersToMetadata(req)

		md := req.Metadata()[fasthttp.HeaderXForwardedFor]
		assert.Equal(t, "originalXForwardedFor", md.Values[0])
		assert.Equal(t, "1", md.Values[1])

		md = req.Metadata()[fasthttp.HeaderXForwardedHost]
		assert.Equal(t, "originalXForwardedHost", md.Values[0])
		assert.Equal(t, "2", md.Values[1])

		md = req.Metadata()[fasthttp.HeaderForwarded]
		assert.Equal(t, "originalForwarded", md.Values[0])
		assert.Equal(t, "for=1;by=1;host=2", md.Values[1])
	})
}

func TestKubernetesNamespace(t *testing.T) {
	t.Run("no namespace", func(t *testing.T) {
		appID := "app1"

		dm := newDirectMessaging()
		id, ns, err := dm.requestAppIDAndNamespace(appID)

		assert.NoError(t, err)
		assert.Empty(t, ns)
		assert.Equal(t, appID, id)
	})

	t.Run("with namespace", func(t *testing.T) {
		appID := "app1.ns1"

		dm := newDirectMessaging()
		id, ns, err := dm.requestAppIDAndNamespace(appID)

		assert.NoError(t, err)
		assert.Equal(t, "ns1", ns)
		assert.Equal(t, "app1", id)
	})

	t.Run("invalid namespace", func(t *testing.T) {
		appID := "app1.ns1.ns2"

		dm := newDirectMessaging()
		_, _, err := dm.requestAppIDAndNamespace(appID)

		assert.Error(t, err)
	})
}
