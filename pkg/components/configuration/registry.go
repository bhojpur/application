package configuration

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
	"github.com/bhojpur/service/pkg/configuration"
)

type Configuration struct {
	Name          string
	FactoryMethod func() configuration.Store
}

func New(name string, factoryMethod func() configuration.Store) Configuration {
	return Configuration{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// Registry is an interface for a component that returns registered configuration store implementations.
type Registry interface {
	Register(components ...Configuration)
	Create(name, version string) (configuration.Store, error)
}

type configurationStoreRegistry struct {
	configurationStores map[string]func() configuration.Store
}

// NewRegistry is used to create configuration store registry.
func NewRegistry() Registry {
	return &configurationStoreRegistry{
		configurationStores: map[string]func() configuration.Store{},
	}
}

// Register registers a new factory method that creates an instance of a ConfigurationStore.
// The key is the name of the state store, eg. redis.
func (s *configurationStoreRegistry) Register(components ...Configuration) {
	for _, component := range components {
		s.configurationStores[createFullName(component.Name)] = component.FactoryMethod
	}
}

func (s *configurationStoreRegistry) Create(name, version string) (configuration.Store, error) {
	if method, ok := s.getConfigurationStore(name, version); ok {
		return method(), nil
	}
	return nil, errors.Errorf("couldn't find configuration store %s/%s", name, version)
}

func (s *configurationStoreRegistry) getConfigurationStore(name, version string) (func() configuration.Store, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	configurationStoreFn, ok := s.configurationStores[nameLower+"/"+versionLower]
	if ok {
		return configurationStoreFn, true
	}
	if components.IsInitialVersion(versionLower) {
		configurationStoreFn, ok = s.configurationStores[nameLower]
	}
	return configurationStoreFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("configuration." + name)
}
