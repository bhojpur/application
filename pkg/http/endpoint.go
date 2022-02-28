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

import "github.com/valyala/fasthttp"

// Endpoint is a collection of route information for a Bhojpur Application API.
//
// If an Alias, e.g. "hello", is provided along with the Route, e.g.
// "invoke/app-id/method/hello" and the Version, "v1.0", then two endpoints will
// be installed instead of one. Besiding the canonical Bhojpur Application runtime
// API URL "/v1.0/invoke/app-id/method/hello", one another URL "/hello" is provided
// for the Alias. When Alias URL is used, extra infos are required to pass through
// HTTP headers. For example, application's ID.
type Endpoint struct {
	Methods           []string
	Route             string
	Version           string
	Alias             string
	KeepParamUnescape bool // keep the param in path unescaped
	Handler           fasthttp.RequestHandler
}
