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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClusterDomain(t *testing.T) {
	testCases := []struct {
		content  string
		expected string
	}{
		{
			content:  "search svc.cluster.local #test comment",
			expected: "svc.cluster.local",
		},
		{
			content:  "search default.svc.cluster.local svc.cluster.local cluster.local",
			expected: "cluster.local",
		},
		{
			content:  "",
			expected: "cluster.local",
		},
	}
	for _, tc := range testCases {
		domain, err := getClusterDomain([]byte(tc.content))
		if err != nil {
			t.Fatalf("get kube cluster domain error:%s", err)
		}
		assert.Equal(t, domain, tc.expected)
	}
}

func TestGetSearchDomains(t *testing.T) {
	testCases := []struct {
		content  string
		expected []string
	}{
		{
			content:  "search svc.cluster.local #test comment",
			expected: []string{"svc.cluster.local"},
		},
		{
			content:  "search default.svc.cluster.local svc.cluster.local cluster.local",
			expected: []string{"default.svc.cluster.local", "svc.cluster.local", "cluster.local"},
		},
		{
			content:  "",
			expected: []string{},
		},
	}
	for _, tc := range testCases {
		domains := getResolvSearchDomains([]byte(tc.content))
		assert.ElementsMatch(t, domains, tc.expected)
	}
}
