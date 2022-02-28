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
	"strings"
)

/* #nosec. */
const (
	// APITokenEnvVar is the environment variable for the api token.
	APITokenEnvVar    = "SVC_API_TOKEN"
	AppAPITokenEnvVar = "APP_API_TOKEN"
	// APITokenHeader is header name for http/gRPC calls to hold the token.
	APITokenHeader = "app-api-token"
)

var excludedRoutes = []string{"/healthz"}

// GetAPIToken returns the value of the api token from an environment variable.
func GetAPIToken() string {
	return os.Getenv(APITokenEnvVar)
}

// GetAppToken returns the value of the app api token from an environment variable.
func GetAppToken() string {
	return os.Getenv(AppAPITokenEnvVar)
}

// ExcludedRoute returns whether a given route should be excluded from a token check.
func ExcludedRoute(route string) bool {
	for _, r := range excludedRoutes {
		if strings.Contains(route, r) {
			return true
		}
	}
	return false
}
