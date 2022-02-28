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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	kauthapi "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
)

func TestValidate(t *testing.T) {
	t.Run("invalid token", func(t *testing.T) {
		fakeClient := &fake.Clientset{}
		fakeClient.Fake.AddReactor(
			"create",
			"tokenreviews",
			func(action core.Action) (bool, runtime.Object, error) {
				return true, &kauthapi.TokenReview{Status: kauthapi.TokenReviewStatus{Error: "bad token"}}, nil
			})

		v := validator{
			client: fakeClient,
			auth:   fakeClient.AuthenticationV1(),
		}

		err := v.Validate("a1:ns1", "a2:ns2", "ns2")
		assert.Equal(t, errors.Errorf("%s: invalid token: bad token", errPrefix).Error(), err.Error())
	})

	t.Run("unauthenticated", func(t *testing.T) {
		fakeClient := &fake.Clientset{}
		fakeClient.Fake.AddReactor(
			"create",
			"tokenreviews",
			func(action core.Action) (bool, runtime.Object, error) {
				return true, &kauthapi.TokenReview{Status: kauthapi.TokenReviewStatus{Authenticated: false}}, nil
			})

		v := validator{
			client: fakeClient,
			auth:   fakeClient.AuthenticationV1(),
		}

		err := v.Validate("a1:ns1", "a2:ns2", "ns")
		expectedErr := errors.Errorf("%s: authentication failed", errPrefix)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("bad token structure", func(t *testing.T) {
		fakeClient := &fake.Clientset{}
		fakeClient.Fake.AddReactor(
			"create",
			"tokenreviews",
			func(action core.Action) (bool, runtime.Object, error) {
				return true, &kauthapi.TokenReview{Status: kauthapi.TokenReviewStatus{Authenticated: true, User: kauthapi.UserInfo{Username: "name"}}}, nil
			})

		v := validator{
			client: fakeClient,
			auth:   fakeClient.AuthenticationV1(),
		}

		err := v.Validate("a1:ns1", "a2:ns2", "ns2")
		expectedErr := errors.Errorf("%s: provided token is not a properly structured service account token", errPrefix)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("token id mismatch", func(t *testing.T) {
		fakeClient := &fake.Clientset{}
		fakeClient.Fake.AddReactor(
			"create",
			"tokenreviews",
			func(action core.Action) (bool, runtime.Object, error) {
				return true, &kauthapi.TokenReview{Status: kauthapi.TokenReviewStatus{Authenticated: true, User: kauthapi.UserInfo{Username: "system:serviceaccount:ns1:a1"}}}, nil
			})

		v := validator{
			client: fakeClient,
			auth:   fakeClient.AuthenticationV1(),
		}

		err := v.Validate("ns2:a1", "ns2:a2", "ns1")
		expectedErr := errors.Errorf("%s: token/id mismatch. received id: ns2:a1", errPrefix)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("empty token", func(t *testing.T) {
		fakeClient := fake.NewSimpleClientset()
		v := validator{
			client: fakeClient,
			auth:   fakeClient.AuthenticationV1(),
		}

		err := v.Validate("a1:ns1", "", "ns")
		expectedErr := errors.Errorf("%s: token field in request must not be empty", errPrefix)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("empty id", func(t *testing.T) {
		fakeClient := fake.NewSimpleClientset()
		v := validator{
			client: fakeClient,
			auth:   fakeClient.AuthenticationV1(),
		}

		err := v.Validate("", "a1:ns1", "ns")
		expectedErr := errors.Errorf("%s: id field in request must not be empty", errPrefix)
		assert.Equal(t, expectedErr.Error(), err.Error())
	})

	t.Run("valid authentication", func(t *testing.T) {
		fakeClient := &fake.Clientset{}
		fakeClient.Fake.AddReactor(
			"create",
			"tokenreviews",
			func(action core.Action) (bool, runtime.Object, error) {
				return true, &kauthapi.TokenReview{Status: kauthapi.TokenReviewStatus{Authenticated: true, User: kauthapi.UserInfo{Username: "system:serviceaccount:ns1:a1"}}}, nil
			})

		v := validator{
			client: fakeClient,
			auth:   fakeClient.AuthenticationV1(),
		}

		err := v.Validate("ns1:a1", "ns1:a1", "ns1")
		assert.NoError(t, err)
	})
}
