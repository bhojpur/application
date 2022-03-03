package secretstores

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

	"github.com/bhojpur/service/pkg/secretstores"

	"github.com/bhojpur/application/pkg/components"
)

type (
	// SecretStore is a secret store component definition.
	SecretStore struct {
		Name          string
		FactoryMethod func() secretstores.SecretStore
	}

	// Registry is used to get registered secret store implementations.
	Registry interface {
		Register(components ...SecretStore)
		Create(name, version string) (secretstores.SecretStore, error)
	}

	secretStoreRegistry struct {
		secretStores map[string]func() secretstores.SecretStore
	}
)

// New creates a SecretStore.
func New(name string, factoryMethod func() secretstores.SecretStore) SecretStore {
	return SecretStore{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// NewRegistry returns a new secret store registry.
func NewRegistry() Registry {
	return &secretStoreRegistry{
		secretStores: map[string]func() secretstores.SecretStore{},
	}
}

// Register adds one or many new secret stores to the registry.
func (s *secretStoreRegistry) Register(components ...SecretStore) {
	for _, component := range components {
		s.secretStores[createFullName(component.Name)] = component.FactoryMethod
	}
}

// Create instantiates a secret store based on `name`.
func (s *secretStoreRegistry) Create(name, version string) (secretstores.SecretStore, error) {
	if method, ok := s.getSecretStore(name, version); ok {
		return method(), nil
	}

	return nil, errors.Errorf("couldn't find Bhojpur Application runtime secret store %s/%s", name, version)
}

func (s *secretStoreRegistry) getSecretStore(name, version string) (func() secretstores.SecretStore, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	secretStoreFn, ok := s.secretStores[nameLower+"/"+versionLower]
	if ok {
		return secretStoreFn, true
	}
	if components.IsInitialVersion(versionLower) {
		secretStoreFn, ok = s.secretStores[nameLower]
	}
	return secretStoreFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("secretstores." + name)
}
