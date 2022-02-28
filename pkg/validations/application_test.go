package validations

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

	"github.com/stretchr/testify/assert"
)

func TestValidationForKubernetes(t *testing.T) {
	t.Run("invalid length", func(t *testing.T) {
		id := ""
		for i := 0; i < 64; i++ {
			id += "a"
		}
		err := ValidateKubernetesAppID(id)
		assert.Error(t, err)
	})

	t.Run("invalid length if suffix -app is appended", func(t *testing.T) {
		// service name id+"-app" exceeds 63 characters (59 + 5 = 64)
		id := strings.Repeat("a", 59)
		err := ValidateKubernetesAppID(id)
		assert.Error(t, err)
	})

	t.Run("valid id", func(t *testing.T) {
		id := "my-app-id"
		err := ValidateKubernetesAppID(id)
		assert.NoError(t, err)
	})

	t.Run("invalid char: .", func(t *testing.T) {
		id := "my-app-id.app"
		err := ValidateKubernetesAppID(id)
		assert.Error(t, err)
	})

	t.Run("invalid chars space", func(t *testing.T) {
		id := "my-app-id app"
		err := ValidateKubernetesAppID(id)
		assert.Error(t, err)
	})

	t.Run("invalid empty", func(t *testing.T) {
		id := ""
		err := ValidateKubernetesAppID(id)
		assert.Regexp(t, "value for the bhojpur.net/app-id annotation is empty", err.Error())
	})
}
