package pubsub

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

	"github.com/bhojpur/service/pkg/pubsub"

	"github.com/bhojpur/application/pkg/components"
)

type (
	// PubSub is a pub/sub component definition.
	PubSub struct {
		Name          string
		FactoryMethod func() pubsub.PubSub
	}

	// Registry is the interface for callers to get registered pub-sub components.
	Registry interface {
		Register(components ...PubSub)
		Create(name, version string) (pubsub.PubSub, error)
	}

	pubSubRegistry struct {
		messageBuses map[string]func() pubsub.PubSub
	}
)

// New creates a PubSub.
func New(name string, factoryMethod func() pubsub.PubSub) PubSub {
	return PubSub{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// NewRegistry returns a new pub sub registry.
func NewRegistry() Registry {
	return &pubSubRegistry{
		messageBuses: map[string]func() pubsub.PubSub{},
	}
}

// Register registers one or more new message buses.
func (p *pubSubRegistry) Register(components ...PubSub) {
	for _, component := range components {
		p.messageBuses[createFullName(component.Name)] = component.FactoryMethod
	}
}

// Create instantiates a pub/sub based on `name`.
func (p *pubSubRegistry) Create(name, version string) (pubsub.PubSub, error) {
	if method, ok := p.getPubSub(name, version); ok {
		return method(), nil
	}
	return nil, errors.Errorf("couldn't find message bus %s/%s", name, version)
}

func (p *pubSubRegistry) getPubSub(name, version string) (func() pubsub.PubSub, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	pubSubFn, ok := p.messageBuses[nameLower+"/"+versionLower]
	if ok {
		return pubSubFn, true
	}
	if components.IsInitialVersion(versionLower) {
		pubSubFn, ok = p.messageBuses[nameLower]
	}
	return pubSubFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("pubsub." + name)
}
