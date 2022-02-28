package http_test

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
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	h "github.com/bhojpur/service/pkg/middleware"

	"github.com/bhojpur/application/pkg/components/middleware/http"
	http_middleware "github.com/bhojpur/application/pkg/middleware/http"
)

func TestRegistry(t *testing.T) {
	testRegistry := http.NewRegistry()

	t.Run("middleware is registered", func(t *testing.T) {
		const (
			middlewareName   = "mockMiddleware"
			middlewareNameV2 = "mockMiddleware/v2"
			componentName    = "middleware.http." + middlewareName
		)

		// Initiate mock object
		mock := http_middleware.Middleware(func(h fasthttp.RequestHandler) fasthttp.RequestHandler {
			return nil
		})
		mockV2 := http_middleware.Middleware(func(h fasthttp.RequestHandler) fasthttp.RequestHandler {
			return nil
		})
		metadata := h.Metadata{}

		// act
		testRegistry.Register(http.New(middlewareName, func(h.Metadata) (http_middleware.Middleware, error) {
			return mock, nil
		}))
		testRegistry.Register(http.New(middlewareNameV2, func(h.Metadata) (http_middleware.Middleware, error) {
			return mockV2, nil
		}))

		// Function values are not comparable.
		// You can't take the address of a function, but if you print it with
		// the fmt package, it prints its address. So you can use fmt.Sprintf()
		// to get the address of a function value.

		// assert v0 and v1
		p, e := testRegistry.Create(componentName, "v0", metadata)
		assert.NoError(t, e)
		assert.Equal(t, fmt.Sprintf("%v", mock), fmt.Sprintf("%v", p))
		p, e = testRegistry.Create(componentName, "v1", metadata)
		assert.NoError(t, e)
		assert.Equal(t, fmt.Sprintf("%v", mock), fmt.Sprintf("%v", p))

		// assert v2
		pV2, e := testRegistry.Create(componentName, "v2", metadata)
		assert.NoError(t, e)
		assert.Equal(t, fmt.Sprintf("%v", mockV2), fmt.Sprintf("%v", pV2))

		// check case-insensitivity
		pV2, e = testRegistry.Create(strings.ToUpper(componentName), "V2", metadata)
		assert.NoError(t, e)
		assert.Equal(t, fmt.Sprintf("%v", mockV2), fmt.Sprintf("%v", pV2))
	})

	t.Run("middleware is not registered", func(t *testing.T) {
		const (
			middlewareName = "fakeMiddleware"
			componentName  = "middleware.http." + middlewareName
		)

		metadata := h.Metadata{}

		// act
		p, actualError := testRegistry.Create(componentName, "v1", metadata)
		expectedError := errors.Errorf("HTTP middleware %s/v1 has not been registered", componentName)

		// assert
		assert.Nil(t, p)
		assert.Equal(t, expectedError.Error(), actualError.Error())
	})
}
