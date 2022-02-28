package security

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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIToken(t *testing.T) {
	t.Run("existing token", func(t *testing.T) {
		/* #nosec */
		token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1OTA1NTQ1NzMsImV4cCI6MTYyMjA5MDU3MywiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSIsIkdpdmVuTmFtZSI6IkpvaG5ueSIsIlN1cm5hbWUiOiJSb2NrZXQiLCJFbWFpbCI6Impyb2NrZXRAZXhhbXBsZS5jb20iLCJSb2xlIjpbIk1hbmFnZXIiLCJQcm9qZWN0IEFkbWluaXN0cmF0b3IiXX0.QLFl8ZqC48DOsT7SmXA794nivmqGgylzjrUu6JhXPW4"
		os.Setenv(APITokenEnvVar, token)
		defer os.Unsetenv(APITokenEnvVar)

		apitoken := GetAPIToken()
		assert.Equal(t, token, apitoken)
	})

	t.Run("non-existent token", func(t *testing.T) {
		token := GetAPIToken()
		assert.Equal(t, "", token)
	})
}

func TestAppToken(t *testing.T) {
	t.Run("existing token", func(t *testing.T) {
		/* #nosec */
		token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1OTA1NTQ1NzMsImV4cCI6MTYyMjA5MDU3MywiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSIsIkdpdmVuTmFtZSI6IkpvaG5ueSIsIlN1cm5hbWUiOiJSb2NrZXQiLCJFbWFpbCI6Impyb2NrZXRAZXhhbXBsZS5jb20iLCJSb2xlIjpbIk1hbmFnZXIiLCJQcm9qZWN0IEFkbWluaXN0cmF0b3IiXX0.QLFl8ZqC48DOsT7SmXA794nivmqGgylzjrUu6JhXPW4"
		os.Setenv(AppAPITokenEnvVar, token)
		defer os.Unsetenv(AppAPITokenEnvVar)

		apitoken := GetAppToken()
		assert.Equal(t, token, apitoken)
	})

	t.Run("non-existent token", func(t *testing.T) {
		token := GetAppToken()
		assert.Equal(t, "", token)
	})
}

func TestExcludedRoute(t *testing.T) {
	t.Run("healthz route is excluded", func(t *testing.T) {
		route := "v1.0/healthz"
		excluded := ExcludedRoute(route)
		assert.True(t, excluded)
	})

	t.Run("custom route is not excluded", func(t *testing.T) {
		route := "v1.0/state"
		excluded := ExcludedRoute(route)
		assert.False(t, excluded)
	})
}
