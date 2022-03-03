package nameresolution

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

	nr "github.com/bhojpur/service/pkg/nameresolution"

	"github.com/bhojpur/application/pkg/components"
)

type (
	// NameResolution is a name resolution component definition.
	NameResolution struct {
		Name          string
		FactoryMethod func() nr.Resolver
	}

	// Registry handles registering and creating name resolution components.
	Registry interface {
		Register(components ...NameResolution)
		Create(name, version string) (nr.Resolver, error)
	}

	nameResolutionRegistry struct {
		resolvers map[string]func() nr.Resolver
	}
)

// New creates a NameResolution.
func New(name string, factoryMethod func() nr.Resolver) NameResolution {
	return NameResolution{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// NewRegistry creates a name resolution registry.
func NewRegistry() Registry {
	return &nameResolutionRegistry{
		resolvers: map[string]func() nr.Resolver{},
	}
}

// Register adds one or many name resolution components to the registry.
func (s *nameResolutionRegistry) Register(components ...NameResolution) {
	for _, component := range components {
		s.resolvers[createFullName(component.Name)] = component.FactoryMethod
	}
}

// Create instantiates a name resolution resolver based on `name`.
func (s *nameResolutionRegistry) Create(name, version string) (nr.Resolver, error) {
	if method, ok := s.getResolver(createFullName(name), version); ok {
		return method(), nil
	}
	return nil, errors.Errorf("couldn't find Bhojpur Application runtime name resolver %s/%s", name, version)
}

func (s *nameResolutionRegistry) getResolver(name, version string) (func() nr.Resolver, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	resolverFn, ok := s.resolvers[nameLower+"/"+versionLower]
	if ok {
		return resolverFn, true
	}
	if components.IsInitialVersion(versionLower) {
		resolverFn, ok = s.resolvers[nameLower]
	}
	return resolverFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("nameresolution." + name)
}
