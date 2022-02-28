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
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kauthapi "k8s.io/api/authentication/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	kauth "k8s.io/client-go/kubernetes/typed/authentication/v1"

	"github.com/bhojpur/application/pkg/sentry/identity"
)

const (
	errPrefix = "csr validation failed"
)

func NewValidator(client k8s.Interface) identity.Validator {
	return &validator{
		client: client,
		auth:   client.AuthenticationV1(),
	}
}

type validator struct {
	client k8s.Interface
	auth   kauth.AuthenticationV1Interface
}

func (v *validator) Validate(id, token, namespace string) error {
	if id == "" {
		return errors.Errorf("%s: id field in request must not be empty", errPrefix)
	}
	if token == "" {
		return errors.Errorf("%s: token field in request must not be empty", errPrefix)
	}

	review, err := v.auth.TokenReviews().Create(context.TODO(), &kauthapi.TokenReview{Spec: kauthapi.TokenReviewSpec{Token: token}}, v1.CreateOptions{})
	if err != nil {
		return err
	}

	if review.Status.Error != "" {
		return errors.Errorf("%s: invalid token: %s", errPrefix, review.Status.Error)
	}
	if !review.Status.Authenticated {
		return errors.Errorf("%s: authentication failed", errPrefix)
	}

	prts := strings.Split(review.Status.User.Username, ":")
	if len(prts) != 4 || prts[0] != "system" {
		return errors.Errorf("%s: provided token is not a properly structured service account token", errPrefix)
	}

	podSa := prts[3]
	podNs := prts[2]

	if namespace != "" {
		if podNs != namespace {
			return errors.Errorf("%s: namespace mismatch. received namespace: %s", errPrefix, namespace)
		}
	}

	if id != fmt.Sprintf("%s:%s", podNs, podSa) {
		return errors.Errorf("%s: token/id mismatch. received id: %s", errPrefix, id)
	}
	return nil
}
