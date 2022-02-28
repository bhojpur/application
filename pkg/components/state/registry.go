package state

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

	"github.com/bhojpur/application/pkg/components"
	"github.com/bhojpur/service/pkg/state"
)

type State struct {
	Name          string
	FactoryMethod func() state.Store
}

func New(name string, factoryMethod func() state.Store) State {
	return State{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// Registry is an interface for a component that returns registered state store implementations.
type Registry interface {
	Register(components ...State)
	Create(name, version string) (state.Store, error)
}

type stateStoreRegistry struct {
	stateStores map[string]func() state.Store
}

// NewRegistry is used to create state store registry.
func NewRegistry() Registry {
	return &stateStoreRegistry{
		stateStores: map[string]func() state.Store{},
	}
}

// // Register registers a new factory method that creates an instance of a StateStore.
// // The key is the name of the state store, eg. redis.
func (s *stateStoreRegistry) Register(components ...State) {
	for _, component := range components {
		s.stateStores[createFullName(component.Name)] = component.FactoryMethod
	}
}

func (s *stateStoreRegistry) Create(name, version string) (state.Store, error) {
	if method, ok := s.getStateStore(name, version); ok {
		return method(), nil
	}
	return nil, errors.Errorf("couldn't find state store %s/%s", name, version)
}

func (s *stateStoreRegistry) getStateStore(name, version string) (func() state.Store, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	stateStoreFn, ok := s.stateStores[nameLower+"/"+versionLower]
	if ok {
		return stateStoreFn, true
	}
	if components.IsInitialVersion(versionLower) {
		stateStoreFn, ok = s.stateStores[nameLower]
	}
	return stateStoreFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("state." + name)
}
