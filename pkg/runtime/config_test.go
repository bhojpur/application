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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	publicPort := DefaultAppPublicPort
	c := NewRuntimeConfig("app1", []string{"localhost:5050"}, "localhost:5051", "*", "config", "components", "http", "kubernetes",
		3500, 50002, 50001, []string{"1.2.3.4"}, &publicPort, 8080, 7070, true, 1, true, "localhost:5052", true, 4, "", 4, true, time.Second)

	assert.Equal(t, "app1", c.ID)
	assert.Equal(t, "localhost:5050", c.PlacementAddresses[0])
	assert.Equal(t, "localhost:5051", c.Kubernetes.ControlPlaneAddress)
	assert.Equal(t, "*", c.AllowedOrigins)
	assert.Equal(t, "config", c.GlobalConfig)
	assert.Equal(t, "components", c.Standalone.ComponentsPath)
	assert.Equal(t, "http", string(c.ApplicationProtocol))
	assert.Equal(t, "kubernetes", string(c.Mode))
	assert.Equal(t, 3500, c.HTTPPort)
	assert.Equal(t, 50002, c.InternalGRPCPort)
	assert.Equal(t, 50001, c.APIGRPCPort)
	assert.Equal(t, &publicPort, c.PublicPort)
	assert.Equal(t, "1.2.3.4", c.APIListenAddresses[0])
	assert.Equal(t, 8080, c.ApplicationPort)
	assert.Equal(t, 7070, c.ProfilePort)
	assert.Equal(t, true, c.EnableProfiling)
	assert.Equal(t, 1, c.MaxConcurrency)
	assert.Equal(t, true, c.mtlsEnabled)
	assert.Equal(t, "localhost:5052", c.SentryServiceAddress)
	assert.Equal(t, true, c.AppSSL)
	assert.Equal(t, 4, c.MaxRequestBodySize)
	assert.Equal(t, "", c.UnixDomainSocket)
	assert.Equal(t, 4, c.ReadBufferSize)
	assert.Equal(t, true, c.StreamRequestBody)
	assert.Equal(t, time.Second, c.GracefulShutdownDuration)
}
