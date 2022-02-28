package injector

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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInectorConfig(t *testing.T) {
	t.Run("with kube cluster domain env", func(t *testing.T) {
		os.Setenv("TLS_CERT_FILE", "test-cert-file")
		os.Setenv("TLS_KEY_FILE", "test-key-file")
		os.Setenv("SIDECAR_IMAGE", "app-test-image")
		os.Setenv("SIDECAR_IMAGE_PULL_POLICY", "Always")
		os.Setenv("NAMESPACE", "test-namespace")
		os.Setenv("KUBE_CLUSTER_DOMAIN", "cluster.local")
		defer clearenv()

		cfg, err := GetConfig()
		assert.Nil(t, err)
		assert.Equal(t, "test-cert-file", cfg.TLSCertFile)
		assert.Equal(t, "test-key-file", cfg.TLSKeyFile)
		assert.Equal(t, "app-test-image", cfg.SidecarImage)
		assert.Equal(t, "Always", cfg.SidecarImagePullPolicy)
		assert.Equal(t, "test-namespace", cfg.Namespace)
		assert.Equal(t, "cluster.local", cfg.KubeClusterDomain)
	})

	t.Run("not set kube cluster domain env", func(t *testing.T) {
		os.Setenv("TLS_CERT_FILE", "test-cert-file")
		os.Setenv("TLS_KEY_FILE", "test-key-file")
		os.Setenv("SIDECAR_IMAGE", "app-test-image")
		os.Setenv("SIDECAR_IMAGE_PULL_POLICY", "IfNotPresent")
		os.Setenv("NAMESPACE", "test-namespace")
		os.Setenv("KUBE_CLUSTER_DOMAIN", "")
		defer clearenv()

		cfg, err := GetConfig()
		assert.Nil(t, err)
		assert.Equal(t, "test-cert-file", cfg.TLSCertFile)
		assert.Equal(t, "test-key-file", cfg.TLSKeyFile)
		assert.Equal(t, "app-test-image", cfg.SidecarImage)
		assert.Equal(t, "IfNotPresent", cfg.SidecarImagePullPolicy)
		assert.Equal(t, "test-namespace", cfg.Namespace)
		assert.NotEqual(t, "", cfg.KubeClusterDomain)
	})
}

func clearenv() {
	os.Unsetenv("TLS_CERT_FILE")
	os.Unsetenv("TLS_KEY_FILE")
	os.Unsetenv("SIDECAR_IMAGE")
	os.Unsetenv("SIDECAR_IMAGE_PULL_POLICY")
	os.Unsetenv("NAMESPACE")
	os.Unsetenv("KUBE_CLUSTER_DOMAIN")
}
