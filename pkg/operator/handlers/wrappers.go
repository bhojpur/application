package handlers

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
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectWrapper interface {
	GetMatchLabels() map[string]string
	GetTemplateAnnotations() map[string]string
	GetObject() client.Object
}

type DeploymentWrapper struct {
	appsv1.Deployment
}

func (d *DeploymentWrapper) GetMatchLabels() map[string]string {
	return d.Spec.Selector.MatchLabels
}

func (d *DeploymentWrapper) GetTemplateAnnotations() map[string]string {
	return d.Spec.Template.ObjectMeta.Annotations
}

func (d *DeploymentWrapper) GetObject() client.Object {
	return &d.Deployment
}

type StatefulSetWrapper struct {
	appsv1.StatefulSet
}

func (s *StatefulSetWrapper) GetMatchLabels() map[string]string {
	return s.Spec.Selector.MatchLabels
}

func (s *StatefulSetWrapper) GetTemplateAnnotations() map[string]string {
	return s.Spec.Template.ObjectMeta.Annotations
}

func (s *StatefulSetWrapper) GetObject() client.Object {
	return &s.StatefulSet
}
