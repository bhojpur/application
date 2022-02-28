package kubernetes

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
)

func TestHAMode(t *testing.T) {
	t.Run("ha mode", func(t *testing.T) {
		s := []StatusOutput{
			{
				Replicas: 3,
			},
		}

		r := highAvailabilityEnabled(s)
		assert.True(t, r)
	})

	t.Run("non-ha mode", func(t *testing.T) {
		s := []StatusOutput{
			{
				Replicas: 1,
			},
		}

		r := highAvailabilityEnabled(s)
		assert.False(t, r)
	})
}

func TestMTLSChartValues(t *testing.T) {
	val, err := upgradeChartValues("1", "2", "3", true, true, []string{})
	assert.NoError(t, err)
	assert.Len(t, val, 2)
}

func TestArgsChartValues(t *testing.T) {
	val, err := upgradeChartValues("1", "2", "3", true, true, []string{"a=b", "b=c"})
	assert.NoError(t, err)
	assert.Len(t, val, 4)
}

func TestIsDowngrade(t *testing.T) {
	assert.True(t, isDowngrade("1.3.0", "1.4.0-rc.5"))
	assert.True(t, isDowngrade("1.3.0", "1.4.0"))
	assert.False(t, isDowngrade("1.4.0-rc.5", "1.3.0"))
	assert.False(t, isDowngrade("1.4.0", "1.3.0"))
}
