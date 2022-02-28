package identity

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

// Bundle contains all the elements to identify a workload across a trust domain and a namespace.
type Bundle struct {
	ID          string
	Namespace   string
	TrustDomain string
}

// NewBundle returns a new identity bundle.
// namespace and trustDomain are optional parameters. When empty, a nil value is returned.
func NewBundle(id, namespace, trustDomain string) *Bundle {
	// Empty namespace and trust domain result in an empty bundle
	if namespace == "" || trustDomain == "" {
		return nil
	}

	return &Bundle{
		ID:          id,
		Namespace:   namespace,
		TrustDomain: trustDomain,
	}
}
