package components_test

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

	"github.com/bhojpur/application/pkg/components"
)

func TestIsInitialVersion(t *testing.T) {
	tests := map[string]struct {
		version string
		initial bool
	}{
		"empty version":       {version: "", initial: true},
		"unstable":            {version: "v0", initial: true},
		"first stable":        {version: "v1", initial: true},
		"second stable":       {version: "v2", initial: false},
		"unstable upper":      {version: "V0", initial: true},
		"first stable upper":  {version: "V1", initial: true},
		"second stable upper": {version: "V2", initial: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual := components.IsInitialVersion(tc.version)
			assert.Equal(t, tc.initial, actual)
		})
	}
}
