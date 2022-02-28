package scopes

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

func TestGetAllowedTopics(t *testing.T) {
	allowedTests := []struct {
		Metadata map[string]string
		Target   []string
		Msg      string
	}{
		{
			Metadata: nil,
			Target:   []string{},
			Msg:      "pass",
		},
		{
			Metadata: map[string]string{
				"allowedTopics": "topic1,topic2,topic3",
			},
			Target: []string{"topic1", "topic2", "topic3"},
			Msg:    "pass",
		},
		{
			Metadata: map[string]string{
				"allowedTopics": "topic1, topic2, topic3",
			},
			Target: []string{"topic1", "topic2", "topic3"},
			Msg:    "pass, include whitespace",
		},
		{
			Metadata: map[string]string{
				"allowedTopics": "",
			},
			Target: []string{},
			Msg:    "pass",
		},
		{
			Metadata: map[string]string{
				"allowedTopics": "topic1, topic1, topic1",
			},
			Target: []string{"topic1"},
			Msg:    "pass, include whitespace and repeated topic",
		},
	}
	for _, item := range allowedTests {
		assert.Equal(t, GetAllowedTopics(item.Metadata), item.Target)
	}
}

func TestGetScopedTopics(t *testing.T) {
	scopedTests := []struct {
		Scope    string
		AppID    string
		Metadata map[string]string
		Target   []string
		Msg      string
	}{
		{
			Scope:    "subscriptionScopes",
			AppID:    "appid1",
			Metadata: map[string]string{},
			Target:   []string{},
			Msg:      "pass",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid2=topic1",
			},
			Target: []string{},
			Msg:    "pass",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid1=topic1",
			},
			Target: []string{"topic1"},
			Msg:    "pass",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid1=topic1, topic2",
			},
			Target: []string{"topic1", "topic2"},
			Msg:    "pass, include whitespace",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid1=topic1;appid1=topic2",
			},
			Target: []string{"topic1", "topic2"},
			Msg:    "pass, include repeated appid",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid1=topic1;appid1=topic1",
			},
			Target: []string{"topic1"},
			Msg:    "pass, include repeated appid and topic",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid1",
			},
			Target: []string{},
			Msg:    "pass",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid2",
			},
			Target: []string{},
			Msg:    "pass",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid1;appid2=topic2",
			},
			Target: []string{},
			Msg:    "pass",
		},
		{
			Scope: "subscriptionScopes",
			AppID: "appid1",
			Metadata: map[string]string{
				"subscriptionScopes": "appid1=topic1;appid2",
			},
			Target: []string{"topic1"},
			Msg:    "pass",
		},
	}
	for _, item := range scopedTests {
		assert.Equal(t,
			GetScopedTopics(item.Scope, item.AppID, item.Metadata),
			item.Target)
	}
}
