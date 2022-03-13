package identity

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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSPIFFEID(t *testing.T) {
	t.Run("valid arguments", func(t *testing.T) {
		id, err := CreateSPIFFEID("td1", "ns1", "app1")
		assert.NoError(t, err)
		assert.Equal(t, "spiffe://td1/ns/ns1/app1", id)
	})

	t.Run("missing trust domain", func(t *testing.T) {
		id, err := CreateSPIFFEID("", "ns1", "app1")
		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("missing namespace", func(t *testing.T) {
		id, err := CreateSPIFFEID("td1", "", "app1")
		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("missing app id", func(t *testing.T) {
		id, err := CreateSPIFFEID("td1", "ns1", "")
		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("invalid trust domain", func(t *testing.T) {
		id, err := CreateSPIFFEID("td1:a", "ns1", "app1")
		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("invalid trust domain size", func(t *testing.T) {
		td := ""
		for i := 0; i < 255; i++ {
			td = fmt.Sprintf("%s%v", td, i)
		}

		id, err := CreateSPIFFEID(td, "ns1", "app1")
		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("invalid spiffe id size", func(t *testing.T) {
		appID := ""
		for i := 0; i < 2048; i++ {
			appID = fmt.Sprintf("%s%v", appID, i)
		}

		id, err := CreateSPIFFEID("td1", "ns1", appID)
		assert.Error(t, err)
		assert.Empty(t, id)
	})
}
