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
)

const (
	defaultMetricsPort    = "9090"
	defaultMetricsEnabled = true
)

// Options defines the sets of options for Bhojpur Application logging.
type Options struct {
	// OutputLevel is the level of logging
	MetricsEnabled bool

	Port string
}

func defaultMetricOptions() *Options {
	return &Options{
		Port:           defaultMetricsPort,
		MetricsEnabled: defaultMetricsEnabled,
	}
}

// MetricsPort gets metrics port.
func (o *Options) MetricsPort() uint64 {
	port, err := strconv.ParseUint(o.Port, 10, 64)
	if err != nil {
		// Use default metrics port as a fallback
		port, _ = strconv.ParseUint(defaultMetricsPort, 10, 64)
	}

	return port
}

// AttachCmdFlags attaches metrics options to command flags.
func (o *Options) AttachCmdFlags(
	stringVar func(p *string, name string, value string, usage string),
	boolVar func(p *bool, name string, value bool, usage string)) {
	stringVar(
		&o.Port,
		"metrics-port",
		defaultMetricsPort,
		"The port for the Bhojpur Application runtime metrics server")
	boolVar(
		&o.MetricsEnabled,
		"enable-metrics",
		defaultMetricsEnabled,
		"Enable prometheus metric")
}

// AttachCmdFlag attaches single metrics option to command flags.
func (o *Options) AttachCmdFlag(
	stringVar func(p *string, name string, value string, usage string)) {
	stringVar(
		&o.Port,
		"metrics-port",
		defaultMetricsPort,
		"The port for the Bhojpur Application runtime metrics server")
}
