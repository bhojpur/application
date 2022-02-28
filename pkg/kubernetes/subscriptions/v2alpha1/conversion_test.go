package v2alpha1_test

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
	"github.com/stretchr/testify/require"

	"github.com/bhojpur/application/pkg/kubernetes/subscriptions/v1alpha1"
	"github.com/bhojpur/application/pkg/kubernetes/subscriptions/v2alpha1"
)

func TestConversion(t *testing.T) {
	// Test converting to and from v1alpha1
	subscriptionV2 := v2alpha1.Subscription{
		Scopes: []string{"app1", "app2"},
		Spec: v2alpha1.SubscriptionSpec{
			Pubsubname: "testPubSub",
			Topic:      "topicName",
			Metadata: map[string]string{
				"testName": "testValue",
			},
			Routes: v2alpha1.Routes{
				Default: "testPath",
			},
		},
	}

	var subscriptionV1 v1alpha1.Subscription
	err := subscriptionV2.ConvertTo(&subscriptionV1)
	require.NoError(t, err)

	var actual v2alpha1.Subscription
	err = actual.ConvertFrom(&subscriptionV1)
	require.NoError(t, err)

	assert.Equal(t, &subscriptionV2, &actual)
}
