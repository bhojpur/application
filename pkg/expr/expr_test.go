package expr_test

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
	"github.com/stretchr/testify/require"

	"github.com/bhojpur/application/pkg/expr"
)

func TestEval(t *testing.T) {
	var e expr.Expr
	code := `(has(input.test) && input.test == 1234) || (has(result.test) && result.test == 5678)`
	err := e.DecodeString(code)
	require.NoError(t, err)
	assert.Equal(t, code, e.String())
	result, err := e.Eval(map[string]interface{}{
		"input": map[string]interface{}{
			"test": 1234,
		},
		"result": map[string]interface{}{
			"test": 5678,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestJSONMarshal(t *testing.T) {
	var e expr.Expr
	exprBytes := []byte(`"(has(input.test) && input.test == 1234) || (has(result.test) && result.test == 5678)"`)
	err := e.UnmarshalJSON(exprBytes)
	require.NoError(t, err)
	assert.Equal(t, `(has(input.test) && input.test == 1234) || (has(result.test) && result.test == 5678)`, e.Expr())
	_, err = e.MarshalJSON()
	require.NoError(t, err)
}

var result interface{}

func BenchmarkEval(b *testing.B) {
	var e expr.Expr
	err := e.DecodeString(`(has(input.test) && input.test == 1234) || (has(result.test) && result.test == 5678)`)
	require.NoError(b, err)
	data := map[string]interface{}{
		"input": map[string]interface{}{
			"test": 1234,
		},
		"result": map[string]interface{}{
			"test": 5678,
		},
	}
	var r interface{}
	for n := 0; n < b.N; n++ {
		r, _ = e.Eval(data)
	}
	result = r
}
