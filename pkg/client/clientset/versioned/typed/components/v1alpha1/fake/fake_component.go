// Code generated by client-gen. DO NOT EDIT.

package fake

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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"

	v1alpha1 "github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
)

// FakeComponents implements ComponentInterface
type FakeComponents struct {
	Fake *FakeComponentsV1alpha1
	ns   string
}

var componentsResource = schema.GroupVersionResource{Group: "components.bhojpur.net", Version: "v1alpha1", Resource: "components"}

var componentsKind = schema.GroupVersionKind{Group: "components.bhojpur.net", Version: "v1alpha1", Kind: "Component"}

// Get takes name of the component, and returns the corresponding component object, and an error if there is any.
func (c *FakeComponents) Get(name string, options v1.GetOptions) (result *v1alpha1.Component, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(componentsResource, c.ns, name), &v1alpha1.Component{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Component), err
}

// List takes label and field selectors, and returns the list of Components that match those selectors.
func (c *FakeComponents) List(opts v1.ListOptions) (result *v1alpha1.ComponentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(componentsResource, componentsKind, c.ns, opts), &v1alpha1.ComponentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ComponentList{ListMeta: obj.(*v1alpha1.ComponentList).ListMeta}
	for _, item := range obj.(*v1alpha1.ComponentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested components.
func (c *FakeComponents) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(componentsResource, c.ns, opts))
}

// Create takes the representation of a component and creates it.  Returns the server's representation of the component, and an error, if there is any.
func (c *FakeComponents) Create(component *v1alpha1.Component) (result *v1alpha1.Component, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(componentsResource, c.ns, component), &v1alpha1.Component{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Component), err
}

// Update takes the representation of a component and updates it. Returns the server's representation of the component, and an error, if there is any.
func (c *FakeComponents) Update(component *v1alpha1.Component) (result *v1alpha1.Component, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(componentsResource, c.ns, component), &v1alpha1.Component{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Component), err
}

// Delete takes name of the component and deletes it. Returns an error if one occurs.
func (c *FakeComponents) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(componentsResource, c.ns, name), &v1alpha1.Component{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeComponents) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(componentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.ComponentList{})
	return err
}

// Patch applies the patch and returns the patched component.
func (c *FakeComponents) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Component, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(componentsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Component{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Component), err
}