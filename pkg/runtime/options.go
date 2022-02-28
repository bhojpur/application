package runtime

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
	"github.com/bhojpur/application/pkg/components/bindings"
	"github.com/bhojpur/application/pkg/components/configuration"
	"github.com/bhojpur/application/pkg/components/middleware/http"
	"github.com/bhojpur/application/pkg/components/nameresolution"
	"github.com/bhojpur/application/pkg/components/pubsub"
	"github.com/bhojpur/application/pkg/components/secretstores"
	"github.com/bhojpur/application/pkg/components/state"
)

type (
	// runtimeOpts encapsulates the components to include in the runtime.
	runtimeOpts struct {
		secretStores    []secretstores.SecretStore
		states          []state.State
		configurations  []configuration.Configuration
		pubsubs         []pubsub.PubSub
		nameResolutions []nameresolution.NameResolution
		inputBindings   []bindings.InputBinding
		outputBindings  []bindings.OutputBinding
		httpMiddleware  []http.Middleware

		componentsCallback ComponentsCallback
	}

	// Option is a function that customizes the runtime.
	Option func(o *runtimeOpts)
)

// WithSecretStores adds secret store components to the runtime.
func WithSecretStores(secretStores ...secretstores.SecretStore) Option {
	return func(o *runtimeOpts) {
		o.secretStores = append(o.secretStores, secretStores...)
	}
}

// WithStates adds state store components to the runtime.
func WithStates(states ...state.State) Option {
	return func(o *runtimeOpts) {
		o.states = append(o.states, states...)
	}
}

// WithConfigurations adds configuration store components to the runtime.
func WithConfigurations(configurations ...configuration.Configuration) Option {
	return func(o *runtimeOpts) {
		o.configurations = append(o.configurations, configurations...)
	}
}

// WithPubSubs adds pubsub store components to the runtime.
func WithPubSubs(pubsubs ...pubsub.PubSub) Option {
	return func(o *runtimeOpts) {
		o.pubsubs = append(o.pubsubs, pubsubs...)
	}
}

// WithNameResolutions adds name resolution components to the runtime.
func WithNameResolutions(nameResolutions ...nameresolution.NameResolution) Option {
	return func(o *runtimeOpts) {
		o.nameResolutions = append(o.nameResolutions, nameResolutions...)
	}
}

// WithInputBindings adds input binding components to the runtime.
func WithInputBindings(inputBindings ...bindings.InputBinding) Option {
	return func(o *runtimeOpts) {
		o.inputBindings = append(o.inputBindings, inputBindings...)
	}
}

// WithOutputBindings adds output binding components to the runtime.
func WithOutputBindings(outputBindings ...bindings.OutputBinding) Option {
	return func(o *runtimeOpts) {
		o.outputBindings = append(o.outputBindings, outputBindings...)
	}
}

// WithHTTPMiddleware adds HTTP middleware components to the runtime.
func WithHTTPMiddleware(httpMiddleware ...http.Middleware) Option {
	return func(o *runtimeOpts) {
		o.httpMiddleware = append(o.httpMiddleware, httpMiddleware...)
	}
}

// WithComponentsCallback sets the components callback for applications that embed Dapr.
func WithComponentsCallback(componentsCallback ComponentsCallback) Option {
	return func(o *runtimeOpts) {
		o.componentsCallback = componentsCallback
	}
}
