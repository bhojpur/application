package bindings

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
	"strings"

	"github.com/pkg/errors"

	"github.com/bhojpur/service/pkg/bindings"

	"github.com/bhojpur/application/pkg/components"
)

type (
	// InputBinding is an input binding component definition.
	InputBinding struct {
		Name          string
		FactoryMethod func() bindings.InputBinding
	}

	// OutputBinding is an output binding component definition.
	OutputBinding struct {
		Name          string
		FactoryMethod func() bindings.OutputBinding
	}

	// Registry is the interface of a components that allows callers to get registered instances of input and output bindings.
	Registry interface {
		RegisterInputBindings(components ...InputBinding)
		RegisterOutputBindings(components ...OutputBinding)
		HasInputBinding(name, version string) bool
		HasOutputBinding(name, version string) bool
		CreateInputBinding(name, version string) (bindings.InputBinding, error)
		CreateOutputBinding(name, version string) (bindings.OutputBinding, error)
	}

	bindingsRegistry struct {
		inputBindings  map[string]func() bindings.InputBinding
		outputBindings map[string]func() bindings.OutputBinding
	}
)

// NewInput creates a InputBinding.
func NewInput(name string, factoryMethod func() bindings.InputBinding) InputBinding {
	return InputBinding{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// NewOutput creates a OutputBinding.
func NewOutput(name string, factoryMethod func() bindings.OutputBinding) OutputBinding {
	return OutputBinding{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// NewRegistry is used to create new bindings.
func NewRegistry() Registry {
	return &bindingsRegistry{
		inputBindings:  map[string]func() bindings.InputBinding{},
		outputBindings: map[string]func() bindings.OutputBinding{},
	}
}

// RegisterInputBindings registers one or more new input bindings.
func (b *bindingsRegistry) RegisterInputBindings(components ...InputBinding) {
	for _, component := range components {
		b.inputBindings[createFullName(component.Name)] = component.FactoryMethod
	}
}

// RegisterOutputBindings registers one or more new output bindings.
func (b *bindingsRegistry) RegisterOutputBindings(components ...OutputBinding) {
	for _, component := range components {
		b.outputBindings[createFullName(component.Name)] = component.FactoryMethod
	}
}

// CreateInputBinding Create instantiates an input binding based on `name`.
func (b *bindingsRegistry) CreateInputBinding(name, version string) (bindings.InputBinding, error) {
	if method, ok := b.getInputBinding(name, version); ok {
		return method(), nil
	}
	return nil, errors.Errorf("couldn't find input binding %s/%s", name, version)
}

// CreateOutputBinding Create instantiates an output binding based on `name`.
func (b *bindingsRegistry) CreateOutputBinding(name, version string) (bindings.OutputBinding, error) {
	if method, ok := b.getOutputBinding(name, version); ok {
		return method(), nil
	}
	return nil, errors.Errorf("couldn't find output binding %s/%s", name, version)
}

// HasInputBinding checks if an input binding based on `name` exists in the registry.
func (b *bindingsRegistry) HasInputBinding(name, version string) bool {
	_, ok := b.getInputBinding(name, version)
	return ok
}

// HasOutputBinding checks if an output binding based on `name` exists in the registry.
func (b *bindingsRegistry) HasOutputBinding(name, version string) bool {
	_, ok := b.getOutputBinding(name, version)
	return ok
}

func (b *bindingsRegistry) getInputBinding(name, version string) (func() bindings.InputBinding, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	bindingFn, ok := b.inputBindings[nameLower+"/"+versionLower]
	if ok {
		return bindingFn, true
	}
	if components.IsInitialVersion(versionLower) {
		bindingFn, ok = b.inputBindings[nameLower]
	}
	return bindingFn, ok
}

func (b *bindingsRegistry) getOutputBinding(name, version string) (func() bindings.OutputBinding, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	bindingFn, ok := b.outputBindings[nameLower+"/"+versionLower]
	if ok {
		return bindingFn, true
	}
	if components.IsInitialVersion(versionLower) {
		bindingFn, ok = b.outputBindings[nameLower]
	}
	return bindingFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("bindings." + name)
}
