package grpc

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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/bhojpur/application/pkg/config"
)

func TestSetAPIEndpointsMiddlewareUnary(t *testing.T) {
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	t.Run("state endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "state",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["state.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "state.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("publish endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "publish",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["publish.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "publish.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("actors endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "actors",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["actors.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "actors.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("bindings endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "bindings",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["bindings.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "bindings.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("secrets endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "secrets",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["secrets.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "secrets.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("metadata endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "metadata",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["metadata.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "metadata.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("shutdown endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "shutdown",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["shutdown.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "shutdown.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("invoke endpoints allowed", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "invoke",
				Version:  "v1",
				Protocol: "grpc",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, e := range endpoints["invoke.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for k, v := range endpoints {
			if k != "invoke.v1" {
				for _, e := range v {
					_, err := f(nil, nil, &grpc.UnaryServerInfo{
						FullMethod: e,
					}, h)
					assert.Error(t, err)
				}
			}
		}
	})

	t.Run("no rules, all endpoints are allowed", func(t *testing.T) {
		f := setAPIEndpointsMiddlewareUnary(nil)

		for _, e := range endpoints["invoke.v1"] {
			_, err := f(nil, nil, &grpc.UnaryServerInfo{
				FullMethod: e,
			}, h)
			assert.NoError(t, err)
		}

		for _, v := range endpoints {
			for _, e := range v {
				_, err := f(nil, nil, &grpc.UnaryServerInfo{
					FullMethod: e,
				}, h)
				assert.NoError(t, err)
			}
		}
	})

	t.Run("protocol mismatch, rule not applied", func(t *testing.T) {
		a := []config.APIAccessRule{
			{
				Name:     "state",
				Version:  "v1",
				Protocol: "http",
			},
		}

		f := setAPIEndpointsMiddlewareUnary(a)

		for _, v := range endpoints {
			for _, e := range v {
				_, err := f(nil, nil, &grpc.UnaryServerInfo{
					FullMethod: e,
				}, h)
				assert.NoError(t, err)
			}
		}
	})
}
