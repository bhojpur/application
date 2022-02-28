package pubsub

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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCloudEvent(t *testing.T) {
	t.Run("raw payload", func(t *testing.T) {
		ce, err := NewCloudEvent(&CloudEvent{
			ID:              "a",
			Topic:           "b",
			Data:            []byte("hello"),
			Pubsub:          "c",
			DataContentType: "",
			TraceID:         "d",
		})
		assert.NoError(t, err)
		assert.Equal(t, "a", ce["source"].(string))
		assert.Equal(t, "b", ce["topic"].(string))
		assert.Equal(t, "hello", ce["data"].(string))
		assert.Equal(t, "text/plain", ce["datacontenttype"].(string))
		assert.Equal(t, "d", ce["traceid"].(string))
	})

	t.Run("raw payload no data", func(t *testing.T) {
		ce, err := NewCloudEvent(&CloudEvent{
			ID:              "a",
			Topic:           "b",
			Pubsub:          "c",
			DataContentType: "",
			TraceID:         "d",
		})
		assert.NoError(t, err)
		assert.Equal(t, "a", ce["source"].(string))
		assert.Equal(t, "b", ce["topic"].(string))
		assert.Empty(t, ce["data"])
		assert.Equal(t, "text/plain", ce["datacontenttype"].(string))
		assert.Equal(t, "d", ce["traceid"].(string))
	})

	t.Run("custom cloudevent", func(t *testing.T) {
		m := map[string]interface{}{
			"specversion":     "1.0",
			"id":              "event",
			"datacontenttype": "text/plain",
			"data":            "world",
		}
		b, _ := json.Marshal(m)

		ce, err := NewCloudEvent(&CloudEvent{
			Data:            b,
			DataContentType: "application/cloudevents+json",
			Topic:           "topic1",
			TraceID:         "trace1",
			Pubsub:          "pubsub",
		})
		assert.NoError(t, err)
		assert.Equal(t, "world", ce["data"].(string))
		assert.Equal(t, "text/plain", ce["datacontenttype"].(string))
		assert.Equal(t, "topic1", ce["topic"].(string))
		assert.Equal(t, "trace1", ce["traceid"].(string))
		assert.Equal(t, "pubsub", ce["pubsubname"].(string))
	})
}
