package v2alpha1

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
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/bhojpur/application/pkg/kubernetes/subscriptions/v1alpha1"
)

// +kubebuilder:docs-gen:collapse=Imports

/*
Our "spoke" versions need to implement the
[`Convertible`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/conversion?tab=doc#Convertible)
interface.  Namely, they'll need `ConvertTo` and `ConvertFrom` methods to convert to/from
the hub version.
*/

/*
ConvertTo is expected to modify its argument to contain the converted object.
Most of the conversion is straightforward copying, except for converting our changed field.
*/
// ConvertTo converts this Subscription to the Hub version (v1).
func (s *Subscription) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1alpha1.Subscription)
	if !ok {
		return errors.New("expected to to convert to *v1alpha1.Subscription")
	}

	// Copy scopes
	dst.Scopes = s.Scopes

	// ObjectMeta
	dst.ObjectMeta = s.ObjectMeta

	// Spec
	dst.Spec.Pubsubname = s.Spec.Pubsubname
	dst.Spec.Topic = s.Spec.Topic
	dst.Spec.Metadata = s.Spec.Metadata
	dst.Spec.Route = s.Spec.Routes.Default

	// +kubebuilder:docs-gen:collapse=rote conversion
	return nil
}

/*
ConvertFrom is expected to modify its receiver to contain the converted object.
Most of the conversion is straightforward copying, except for converting our changed field.
*/
// ConvertFrom converts from the Hub version (v1) to this version.
func (s *Subscription) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1alpha1.Subscription)
	if !ok {
		return errors.New("expected to to convert from *v1alpha1.Subscription")
	}

	// Copy scopes
	s.Scopes = src.Scopes

	// ObjectMeta
	s.ObjectMeta = src.ObjectMeta

	// Spec
	s.Spec.Pubsubname = src.Spec.Pubsubname
	s.Spec.Topic = src.Spec.Topic
	s.Spec.Metadata = src.Spec.Metadata
	s.Spec.Routes.Default = src.Spec.Route

	// +kubebuilder:docs-gen:collapse=rote conversion
	return nil
}
