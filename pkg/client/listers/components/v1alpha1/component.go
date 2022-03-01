// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	v1alpha1 "github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
)

// ComponentLister helps list Components.
type ComponentLister interface {
	// List lists all Components in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.Component, err error)
	// Components returns an object that can list and get Components.
	Components(namespace string) ComponentNamespaceLister
	ComponentListerExpansion
}

// componentLister implements the ComponentLister interface.
type componentLister struct {
	indexer cache.Indexer
}

// NewComponentLister returns a new ComponentLister.
func NewComponentLister(indexer cache.Indexer) ComponentLister {
	return &componentLister{indexer: indexer}
}

// List lists all Components in the indexer.
func (s *componentLister) List(selector labels.Selector) (ret []*v1alpha1.Component, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Component))
	})
	return ret, err
}

// Components returns an object that can list and get Components.
func (s *componentLister) Components(namespace string) ComponentNamespaceLister {
	return componentNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ComponentNamespaceLister helps list and get Components.
type ComponentNamespaceLister interface {
	// List lists all Components in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.Component, err error)
	// Get retrieves the Component from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.Component, error)
	ComponentNamespaceListerExpansion
}

// componentNamespaceLister implements the ComponentNamespaceLister
// interface.
type componentNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Components in the indexer for a given namespace.
func (s componentNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Component, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Component))
	})
	return ret, err
}

// Get retrieves the Component from the indexer for a given namespace and name.
func (s componentNamespaceLister) Get(name string) (*v1alpha1.Component, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("component"), name)
	}
	return obj.(*v1alpha1.Component), nil
}