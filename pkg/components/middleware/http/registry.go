package http

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

	middleware "github.com/bhojpur/service/pkg/middleware"

	"github.com/bhojpur/application/pkg/components"
	http_middleware "github.com/bhojpur/application/pkg/middleware/http"
)

type (
	// Middleware is a HTTP middleware component definition.
	Middleware struct {
		Name          string
		FactoryMethod FactoryMethod
	}

	// Registry is the interface for callers to get registered HTTP middleware.
	Registry interface {
		Register(components ...Middleware)
		Create(name, version string, metadata middleware.Metadata) (http_middleware.Middleware, error)
	}

	httpMiddlewareRegistry struct {
		middleware map[string]FactoryMethod
	}

	// FactoryMethod is the method creating middleware from metadata.
	FactoryMethod func(metadata middleware.Metadata) (http_middleware.Middleware, error)
)

// New creates a Middleware.
func New(name string, factoryMethod FactoryMethod) Middleware {
	return Middleware{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// NewRegistry returns a new HTTP middleware registry.
func NewRegistry() Registry {
	return &httpMiddlewareRegistry{
		middleware: map[string]FactoryMethod{},
	}
}

// Register registers one or more new HTTP middlewares.
func (p *httpMiddlewareRegistry) Register(components ...Middleware) {
	for _, component := range components {
		p.middleware[createFullName(component.Name)] = component.FactoryMethod
	}
}

// Create instantiates a HTTP middleware based on `name`.
func (p *httpMiddlewareRegistry) Create(name, version string, metadata middleware.Metadata) (http_middleware.Middleware, error) {
	if method, ok := p.getMiddleware(name, version); ok {
		mid, err := method(metadata)
		if err != nil {
			return nil, errors.Errorf("error creating Bhojpur Application runtime HTTP middleware %s/%s: %s", name, version, err)
		}
		return mid, nil
	}
	return nil, errors.Errorf("HTTP middleware %s/%s has not been registered", name, version)
}

func (p *httpMiddlewareRegistry) getMiddleware(name, version string) (FactoryMethod, bool) {
	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	middlewareFn, ok := p.middleware[nameLower+"/"+versionLower]
	if ok {
		return middlewareFn, true
	}
	if components.IsInitialVersion(versionLower) {
		middlewareFn, ok = p.middleware[nameLower]
	}
	return middlewareFn, ok
}

func createFullName(name string) string {
	return strings.ToLower("middleware.http." + name)
}
