package testing

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
	mock "github.com/stretchr/testify/mock"

	"github.com/bhojpur/service/pkg/bindings"
)

// MockBinding is a mock input/output component object.
type MockBinding struct {
	mock.Mock
}

// Init is a mock initialization method.
func (m *MockBinding) Init(metadata bindings.Metadata) error {
	return nil
}

// Read is a mock read method.
func (m *MockBinding) Read(handler func(*bindings.ReadResponse) ([]byte, error)) error {
	args := m.Called(handler)
	return args.Error(0)
}

// Invoke is a mock invoke method.
func (m *MockBinding) Invoke(req *bindings.InvokeRequest) (*bindings.InvokeResponse, error) {
	args := m.Called(req)
	return nil, args.Error(0)
}

// Operations is a mock operations method.
func (m *MockBinding) Operations() []bindings.OperationKind {
	return []bindings.OperationKind{bindings.CreateOperation}
}

func (m *MockBinding) Close() error {
	return nil
}
