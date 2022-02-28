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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListPodsInterface(t *testing.T) {
	t.Run("empty list pods", func(t *testing.T) {
		k8s := fake.NewSimpleClientset()
		output, err := ListPodsInterface(k8s, map[string]string{
			"test": "test",
		})
		assert.Nil(t, err, "unexpected error")
		assert.NotNil(t, output, "Expected empty list")
		assert.Equal(t, 0, len(output.Items), "Expected length 0")
	})
	t.Run("one matching pod", func(t *testing.T) {
		k8s := fake.NewSimpleClientset((&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test",
				Namespace:   "test",
				Annotations: map[string]string{},
				Labels: map[string]string{
					"test": "test",
				},
			},
		}))
		output, err := ListPodsInterface(k8s, map[string]string{
			"test": "test",
		})
		assert.Nil(t, err, "unexpected error")
		assert.NotNil(t, output, "Expected non empty list")
		assert.Equal(t, 1, len(output.Items), "Expected length 0")
		assert.Equal(t, "test", output.Items[0].Name, "expected name to match")
		assert.Equal(t, "test", output.Items[0].Namespace, "expected namespace to match")
	})
}
