package health

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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	t.Run("unhealthy endpoint custom interval 1, failure threshold 2s", func(t *testing.T) {
		start := time.Now()
		ch := StartEndpointHealthCheck("invalid", WithInterval(time.Second*1), WithFailureThreshold(2))
		for {
			healthy := <-ch
			assert.False(t, healthy)

			d := time.Since(start).Seconds()
			assert.True(t, int(d) < 3)
			return
		}
	})

	t.Run("unhealthy endpoint custom interval 1s, failure threshold 1, initial delay 2s", func(t *testing.T) {
		start := time.Now()
		ch := StartEndpointHealthCheck("invalid", WithInterval(time.Second*1), WithFailureThreshold(1), WithInitialDelay(time.Second*2))
		for {
			healthy := <-ch
			assert.False(t, healthy)

			d := time.Since(start).Seconds()
			assert.True(t, int(d) < 3)
			return
		}
	})

	t.Run("unhealthy endpoint custom interval 1s, failure threshold 2, initial delay 2s", func(t *testing.T) {
		start := time.Now()
		ch := StartEndpointHealthCheck("invalid", WithInterval(time.Second*1), WithFailureThreshold(2), WithInitialDelay(time.Second*2))
		for {
			healthy := <-ch
			assert.False(t, healthy)

			d := time.Since(start).Seconds()
			assert.True(t, int(d) < 4)
			return
		}
	})
}

func TestApplyOptions(t *testing.T) {
	t.Run("valid defaults", func(t *testing.T) {
		opts := healthCheckOptions{}
		applyDefaults(&opts)

		assert.Equal(t, opts.failureThreshold, failureThreshold)
		assert.Equal(t, opts.initialDelay, initialDelay)
		assert.Equal(t, opts.interval, interval)
		assert.Equal(t, opts.requestTimeout, requestTimeout)
		assert.Equal(t, opts.successStatusCode, successStatusCode)
	})

	t.Run("valid custom options", func(t *testing.T) {
		opts := healthCheckOptions{}
		applyDefaults(&opts)

		customOpts := []Option{
			WithFailureThreshold(10),
			WithInitialDelay(time.Second * 11),
			WithInterval(time.Second * 12),
			WithRequestTimeout(time.Second * 13),
			WithSuccessStatusCode(201),
		}
		for _, o := range customOpts {
			o(&opts)
		}
		assert.Equal(t, opts.failureThreshold, 10)
		assert.Equal(t, opts.initialDelay, time.Second*11)
		assert.Equal(t, opts.interval, time.Second*12)
		assert.Equal(t, opts.requestTimeout, time.Second*13)
		assert.Equal(t, opts.successStatusCode, 201)
	})
}

type testServer struct {
	statusCode    int
	numberOfCalls int
}

func (t *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.numberOfCalls++
	w.WriteHeader(t.statusCode)
	w.Write([]byte(""))
}

func TestResponses(t *testing.T) {
	t.Run("default success status", func(t *testing.T) {
		server := httptest.NewServer(&testServer{
			statusCode: 200,
		})
		defer server.Close()

		ch := StartEndpointHealthCheck(server.URL, WithInterval(time.Second*1), WithFailureThreshold(1))
		for {
			healthy := <-ch
			assert.True(t, healthy)

			return
		}
	})

	t.Run("custom success status", func(t *testing.T) {
		server := httptest.NewServer(&testServer{
			statusCode: 201,
		})
		defer server.Close()

		ch := StartEndpointHealthCheck(server.URL, WithInterval(time.Second*1), WithFailureThreshold(1), WithSuccessStatusCode(201))
		for {
			healthy := <-ch
			assert.True(t, healthy)

			return
		}
	})

	t.Run("test fail", func(t *testing.T) {
		server := httptest.NewServer(&testServer{
			statusCode: 500,
		})
		defer server.Close()

		ch := StartEndpointHealthCheck(server.URL, WithInterval(time.Second*1), WithFailureThreshold(1))
		for {
			healthy := <-ch
			assert.False(t, healthy)

			return
		}
	})

	t.Run("test app recovery", func(t *testing.T) {
		test := &testServer{
			statusCode: 500,
		}
		server := httptest.NewServer(test)
		defer server.Close()

		ch := StartEndpointHealthCheck(server.URL, WithInterval(time.Second*1), WithFailureThreshold(1))
		count := 0
		for {
			healthy := <-ch
			count++
			if count != 2 {
				assert.False(t, healthy)
				test.statusCode = 200
			} else {
				assert.True(t, healthy)

				return
			}
		}
	})
}
