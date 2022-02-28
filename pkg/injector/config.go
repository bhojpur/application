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
	"github.com/kelseyhightower/envconfig"

	"github.com/bhojpur/application/pkg/utils"
)

// Config represents configuration options for the Bhojpur Application runtime
// Sidecar Injector webhook server.
type Config struct {
	TLSCertFile            string `envconfig:"TLS_CERT_FILE" required:"true"`
	TLSKeyFile             string `envconfig:"TLS_KEY_FILE" required:"true"`
	SidecarImage           string `envconfig:"SIDECAR_IMAGE" required:"true"`
	SidecarImagePullPolicy string `envconfig:"SIDECAR_IMAGE_PULL_POLICY"`
	Namespace              string `envconfig:"NAMESPACE" required:"true"`
	KubeClusterDomain      string `envconfig:"KUBE_CLUSTER_DOMAIN"`
}

// NewConfigWithDefaults returns a Config object with default values already
// applied. Callers are then free to set custom values for the remaining fields
// and/or override default values.
func NewConfigWithDefaults() Config {
	return Config{
		SidecarImagePullPolicy: "Always",
	}
}

// GetConfig returns configuration derived from environment variables.
func GetConfig() (Config, error) {
	// get config from environment variables
	c := NewConfigWithDefaults()
	err := envconfig.Process("", &c)
	if err != nil {
		return c, err
	}

	if c.KubeClusterDomain == "" {
		// auto-detect KubeClusterDomain from resolv.conf file
		clusterDomain, err := utils.GetKubeClusterDomain()
		if err != nil {
			log.Errorf("failed to get clusterDomain err:%s, set default:%s", err, utils.DefaultKubeClusterDomain)
			c.KubeClusterDomain = utils.DefaultKubeClusterDomain
		} else {
			c.KubeClusterDomain = clusterDomain
		}
	}
	return c, nil
}
