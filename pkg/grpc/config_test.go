package grpc

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

func TestServerConfig(t *testing.T) {
	vals := []interface{}{
		"app1",
		"localhost:5050",
		50001,
		"1.2.3.4",
		"default",
		"td1",
		4,
		"",
		4,
	}

	c := NewServerConfig(vals[0].(string), vals[1].(string), vals[2].(int), []string{vals[3].(string)}, vals[4].(string), vals[5].(string), vals[6].(int), vals[7].(string), vals[8].(int))
	assert.Equal(t, vals[0], c.AppID)
	assert.Equal(t, vals[1], c.HostAddress)
	assert.Equal(t, vals[2], c.Port)
	assert.Equal(t, vals[3], c.APIListenAddresses[0])
	assert.Equal(t, vals[4], c.NameSpace)
	assert.Equal(t, vals[5], c.TrustDomain)
	assert.Equal(t, vals[6], c.MaxRequestBodySize)
	assert.Equal(t, vals[8], c.ReadBufferSize)
}
