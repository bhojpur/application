package kubernetes

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

// This can be removed in the future (>= 1.0) when chart versions align to runtime versions.
var chartVersionsMap = map[string]string{
	"0.7.0": "0.4.0",
	"0.7.1": "0.4.1",
	"0.8.0": "0.4.2",
	"0.9.0": "0.4.3",
}

// chartVersion will return the corresponding Helm Chart version for the given runtime version.
// If the specified version is not found, it is assumed that the chart version equals the runtime version.
func chartVersion(runtimeVersion string) string {
	v, ok := chartVersionsMap[runtimeVersion]
	if ok {
		return v
	}
	return runtimeVersion
}
