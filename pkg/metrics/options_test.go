package metrics

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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		o := defaultMetricOptions()
		assert.Equal(t, defaultMetricsPort, o.Port)
		assert.Equal(t, defaultMetricsEnabled, o.MetricsEnabled)
	})

	t.Run("attaching metrics related cmd flags", func(t *testing.T) {
		o := defaultMetricOptions()

		metricsPortAsserted := false
		testStringVarFn := func(p *string, name string, value string, usage string) {
			if name == "metrics-port" && value == defaultMetricsPort {
				metricsPortAsserted = true
			}
		}

		metricsEnabledAsserted := false
		testBoolVarFn := func(p *bool, name string, value bool, usage string) {
			if name == "enable-metrics" && value == defaultMetricsEnabled {
				metricsEnabledAsserted = true
			}
		}

		o.AttachCmdFlags(testStringVarFn, testBoolVarFn)

		// assert
		assert.True(t, metricsPortAsserted)
		assert.True(t, metricsEnabledAsserted)
	})

	t.Run("parse valid port", func(t *testing.T) {
		o := Options{
			Port:           "1010",
			MetricsEnabled: false,
		}

		assert.Equal(t, uint64(1010), o.MetricsPort())
	})

	t.Run("return default port if port is invalid", func(t *testing.T) {
		o := Options{
			Port:           "invalid",
			MetricsEnabled: false,
		}

		defaultPort, _ := strconv.ParseUint(defaultMetricsPort, 10, 64)

		assert.Equal(t, defaultPort, o.MetricsPort())
	})

	t.Run("attaching single metrics related cmd flag", func(t *testing.T) {
		o := defaultMetricOptions()

		metricsPortAsserted := false
		testStringVarFn := func(p *string, name string, value string, usage string) {
			if name == "metrics-port" && value == defaultMetricsPort {
				metricsPortAsserted = true
			}
		}

		o.AttachCmdFlag(testStringVarFn)

		// assert
		assert.True(t, metricsPortAsserted)
	})
}
