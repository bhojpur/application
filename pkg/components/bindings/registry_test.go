package bindings_test

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

	b "github.com/bhojpur/service/pkg/bindings"

	"github.com/bhojpur/application/pkg/components/bindings"
)

type (
	mockInputBinding struct {
		b.InputBinding
	}

	mockOutputBinding struct {
		b.OutputBinding
	}
)

func TestRegistry(t *testing.T) {
	testRegistry := bindings.NewRegistry()

	t.Run("input binding is registered", func(t *testing.T) {
		const (
			inputBindingName   = "mockInputBinding"
			inputBindingNameV2 = "mockInputBinding/v2"
			componentName      = "bindings." + inputBindingName
		)

		// Initiate mock object
		mockInput := &mockInputBinding{}
		mockInputV2 := &mockInputBinding{}

		// act
		testRegistry.RegisterInputBindings(bindings.NewInput(inputBindingName, func() b.InputBinding {
			return mockInput
		}))
		testRegistry.RegisterInputBindings(bindings.NewInput(inputBindingNameV2, func() b.InputBinding {
			return mockInputV2
		}))

		// assert v0 and v1
		assert.True(t, testRegistry.HasInputBinding(componentName, "v0"))
		p, e := testRegistry.CreateInputBinding(componentName, "v0")
		assert.NoError(t, e)
		assert.Same(t, mockInput, p)
		p, e = testRegistry.CreateInputBinding(componentName, "v1")
		assert.NoError(t, e)
		assert.Same(t, mockInput, p)

		// assert v2
		assert.True(t, testRegistry.HasInputBinding(componentName, "v2"))
		pV2, e := testRegistry.CreateInputBinding(componentName, "v2")
		assert.NoError(t, e)
		assert.Same(t, mockInputV2, pV2)

		// check case-insensitivity
		pV2, e = testRegistry.CreateInputBinding(strings.ToUpper(componentName), "V2")
		assert.NoError(t, e)
		assert.Same(t, mockInputV2, pV2)
	})

	t.Run("input binding is not registered", func(t *testing.T) {
		const (
			inputBindingName = "fakeInputBinding"
			componentName    = "bindings." + inputBindingName
		)

		// act
		assert.False(t, testRegistry.HasInputBinding(componentName, "v0"))
		assert.False(t, testRegistry.HasInputBinding(componentName, "v1"))
		assert.False(t, testRegistry.HasInputBinding(componentName, "v2"))
		p, actualError := testRegistry.CreateInputBinding(componentName, "v1")
		expectedError := errors.Errorf("couldn't find input binding %s/v1", componentName)

		// assert
		assert.Nil(t, p)
		assert.Equal(t, expectedError.Error(), actualError.Error())
	})

	t.Run("output binding is registered", func(t *testing.T) {
		const (
			outputBindingName   = "mockInputBinding"
			outputBindingNameV2 = "mockInputBinding/v2"
			componentName       = "bindings." + outputBindingName
		)

		// Initiate mock object
		mockOutput := &mockOutputBinding{}
		mockOutputV2 := &mockOutputBinding{}

		// act
		testRegistry.RegisterOutputBindings(bindings.NewOutput(outputBindingName, func() b.OutputBinding {
			return mockOutput
		}))
		testRegistry.RegisterOutputBindings(bindings.NewOutput(outputBindingNameV2, func() b.OutputBinding {
			return mockOutputV2
		}))

		// assert v0 and v1
		assert.True(t, testRegistry.HasOutputBinding(componentName, "v0"))
		p, e := testRegistry.CreateOutputBinding(componentName, "v0")
		assert.NoError(t, e)
		assert.Same(t, mockOutput, p)
		assert.True(t, testRegistry.HasOutputBinding(componentName, "v1"))
		p, e = testRegistry.CreateOutputBinding(componentName, "v1")
		assert.NoError(t, e)
		assert.Same(t, mockOutput, p)

		// assert v2
		assert.True(t, testRegistry.HasOutputBinding(componentName, "v2"))
		pV2, e := testRegistry.CreateOutputBinding(componentName, "v2")
		assert.NoError(t, e)
		assert.Same(t, mockOutputV2, pV2)
	})

	t.Run("output binding is not registered", func(t *testing.T) {
		const (
			outputBindingName = "fakeOutputBinding"
			componentName     = "bindings." + outputBindingName
		)

		// act
		assert.False(t, testRegistry.HasOutputBinding(componentName, "v0"))
		assert.False(t, testRegistry.HasOutputBinding(componentName, "v1"))
		assert.False(t, testRegistry.HasOutputBinding(componentName, "v2"))
		p, actualError := testRegistry.CreateOutputBinding(componentName, "v1")
		expectedError := errors.Errorf("couldn't find output binding %s/v1", componentName)

		// assert
		assert.Nil(t, p)
		assert.Equal(t, expectedError.Error(), actualError.Error())
	})
}
