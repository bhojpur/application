package config

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

	"github.com/stretchr/testify/assert"

	app_config "github.com/bhojpur/application/pkg/config"
)

func TestConfig(t *testing.T) {
	t.Run("valid FromConfig, empty config name", func(t *testing.T) {
		defaultConfig := getDefaultConfig()
		c, _ := FromConfigName("")

		assert.Equal(t, defaultConfig, c)
	})

	t.Run("valid default config, self hosted", func(t *testing.T) {
		defaultConfig := getDefaultConfig()
		c, _ := getSelfhostedConfig("")

		assert.Equal(t, defaultConfig, c)
	})

	t.Run("parse configuration", func(t *testing.T) {
		appConfig := app_config.Configuration{
			Spec: app_config.ConfigurationSpec{
				MTLSSpec: app_config.MTLSSpec{
					Enabled:          true,
					WorkloadCertTTL:  "5s",
					AllowedClockSkew: "1h",
				},
			},
		}

		defaultConfig := getDefaultConfig()
		conf, err := parseConfiguration(defaultConfig, &appConfig)
		assert.Nil(t, err)
		assert.Equal(t, "5s", conf.WorkloadCertTTL.String())
		assert.Equal(t, "1h0m0s", conf.AllowedClockSkew.String())
	})
}
