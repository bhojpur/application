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
	"strings"
)

const (
	SubscriptionScopes = "subscriptionScopes"
	PublishingScopes   = "publishingScopes"
	AllowedTopics      = "allowedTopics"
	appsSeparator      = ";"
	appSeparator       = "="
	topicSeparator     = ","
)

// GetScopedTopics returns a list of scoped topics for a given application from a Pub/Sub
// Component properties.
func GetScopedTopics(scope, appID string, metadata map[string]string) []string {
	var (
		existM = map[string]struct{}{}
		topics = []string{}
	)

	if val, ok := metadata[scope]; ok && val != "" {
		val = strings.ReplaceAll(val, " ", "")
		apps := strings.Split(val, appsSeparator)
		for _, a := range apps {
			appTopics := strings.Split(a, appSeparator)
			if len(appTopics) < 2 {
				continue
			}

			app := appTopics[0]
			if app != appID {
				continue
			}

			tempTopics := strings.Split(appTopics[1], topicSeparator)
			for _, tempTopic := range tempTopics {
				if _, ok = existM[tempTopic]; !ok {
					existM[tempTopic] = struct{}{}
					topics = append(topics, tempTopic)
				}
			}
		}
	}
	return topics
}

// GetAllowedTopics return the all topics list of params allowedTopics.
func GetAllowedTopics(metadata map[string]string) []string {
	var (
		existM = map[string]struct{}{}
		topics = []string{}
	)

	if val, ok := metadata[AllowedTopics]; ok && val != "" {
		val = strings.ReplaceAll(val, " ", "")
		tempTopics := strings.Split(val, topicSeparator)
		for _, tempTopic := range tempTopics {
			if _, ok = existM[tempTopic]; !ok {
				existM[tempTopic] = struct{}{}
				topics = append(topics, tempTopic)
			}
		}
	}
	return topics
}
