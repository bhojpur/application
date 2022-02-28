package secretstores_test

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

	ss "github.com/bhojpur/service/pkg/secretstores"

	"github.com/bhojpur/application/pkg/components/secretstores"
)

type mockSecretStore struct {
	ss.SecretStore
}

func TestRegistry(t *testing.T) {
	testRegistry := secretstores.NewRegistry()

	t.Run("secret store is registered", func(t *testing.T) {
		const (
			secretStoreName   = "mockSecretStore"
			secretStoreNameV2 = "mockSecretStore/v2"
			componentName     = "secretstores." + secretStoreName
		)

		// Initiate mock object
		mock := &mockSecretStore{}
		mockV2 := &mockSecretStore{}

		// act
		testRegistry.Register(secretstores.New(secretStoreName, func() ss.SecretStore {
			return mock
		}))
		testRegistry.Register(secretstores.New(secretStoreNameV2, func() ss.SecretStore {
			return mockV2
		}))

		// assert v0 and v1
		p, e := testRegistry.Create(componentName, "v0")
		assert.NoError(t, e)
		assert.Same(t, mock, p)
		p, e = testRegistry.Create(componentName, "v1")
		assert.NoError(t, e)
		assert.Same(t, mock, p)

		// assert v2
		pV2, e := testRegistry.Create(componentName, "v2")
		assert.NoError(t, e)
		assert.Same(t, mockV2, pV2)

		// check case-insensitivity
		pV2, e = testRegistry.Create(strings.ToUpper(componentName), "V2")
		assert.NoError(t, e)
		assert.Same(t, mockV2, pV2)
	})

	t.Run("secret store is not registered", func(t *testing.T) {
		const (
			resolverName  = "fakeSecretStore"
			componentName = "secretstores." + resolverName
		)

		// act
		p, actualError := testRegistry.Create(componentName, "v1")
		expectedError := errors.Errorf("couldn't find secret store %s/v1", componentName)

		// assert
		assert.Nil(t, p)
		assert.Equal(t, expectedError.Error(), actualError.Error())
	})
}
