package pubsub

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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	appt "github.com/bhojpur/application/pkg/testing"
	"github.com/bhojpur/service/pkg/pubsub"
)

func TestCreateFullName(t *testing.T) {
	t.Run("create redis pubsub key name", func(t *testing.T) {
		assert.Equal(t, "pubsub.redis", createFullName("redis"))
	})

	t.Run("create kafka pubsub key name", func(t *testing.T) {
		assert.Equal(t, "pubsub.kafka", createFullName("kafka"))
	})

	t.Run("create azure service bus pubsub key name", func(t *testing.T) {
		assert.Equal(t, "pubsub.azure.servicebus", createFullName("azure.servicebus"))
	})

	t.Run("create rabbitmq pubsub key name", func(t *testing.T) {
		assert.Equal(t, "pubsub.rabbitmq", createFullName("rabbitmq"))
	})
}

func TestCreatePubSub(t *testing.T) {
	testRegistry := NewRegistry()

	t.Run("pubsub messagebus is registered", func(t *testing.T) {
		const (
			pubSubName    = "mockPubSub"
			pubSubNameV2  = "mockPubSub/v2"
			componentName = "pubsub." + pubSubName
		)

		// Initiate mock object
		mockPubSub := new(appt.MockPubSub)
		mockPubSubV2 := new(appt.MockPubSub)

		// act
		testRegistry.Register(New(pubSubName, func() pubsub.PubSub {
			return mockPubSub
		}))
		testRegistry.Register(New(pubSubNameV2, func() pubsub.PubSub {
			return mockPubSubV2
		}))

		// assert v0 and v1
		p, e := testRegistry.Create(componentName, "v0")
		assert.NoError(t, e)
		assert.Same(t, mockPubSub, p)

		p, e = testRegistry.Create(componentName, "v1")
		assert.NoError(t, e)
		assert.Same(t, mockPubSub, p)

		// assert v2
		pV2, e := testRegistry.Create(componentName, "v2")
		assert.NoError(t, e)
		assert.Same(t, mockPubSubV2, pV2)

		// check case-insensitivity
		pV2, e = testRegistry.Create(strings.ToUpper(componentName), "V2")
		assert.NoError(t, e)
		assert.Same(t, mockPubSubV2, pV2)
	})

	t.Run("pubsub messagebus is not registered", func(t *testing.T) {
		const PubSubName = "fakePubSub"

		// act
		p, actualError := testRegistry.Create(createFullName(PubSubName), "v1")
		expectedError := errors.Errorf("couldn't find message bus %s/v1", createFullName(PubSubName))
		// assert
		assert.Nil(t, p)
		assert.Equal(t, expectedError.Error(), actualError.Error())
	})
}
