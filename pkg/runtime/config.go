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
	"time"

	config "github.com/bhojpur/application/pkg/config/modes"
	"github.com/bhojpur/application/pkg/credentials"
	"github.com/bhojpur/application/pkg/utils"
)

// Protocol is a communications protocol.
type Protocol string

const (
	// GRPCProtocol is a gRPC communication protocol.
	GRPCProtocol Protocol = "grpc"
	// HTTPProtocol is a HTTP communication protocol.
	HTTPProtocol Protocol = "http"
	// DefaultAppHTTPPort is the default http port for Bhojpur Application runtime.
	DefaultAppHTTPPort = 3500
	// DefaultAppPublicPort is the default http port for Bhojpur Application runtime.
	DefaultAppPublicPort = 3501
	// DefaultAppAPIGRPCPort is the default API gRPC port for Bhojpur Application runtime.
	DefaultAppAPIGRPCPort = 50001
	// DefaultProfilePort is the default port for profiling endpoints.
	DefaultProfilePort = 7777
	// DefaultMetricsPort is the default port for metrics endpoints.
	DefaultMetricsPort = 9090
	// DefaultMaxRequestBodySize is the default option for the maximum body size in MB for Bhojpur Application runtime HTTP servers.
	DefaultMaxRequestBodySize = 4
	// DefaultAPIListenAddress is which address to listen for the Bhojpur Application runtime HTTP and gRPC APIs. Empty string is all addresses.
	DefaultAPIListenAddress = ""
	// DefaultReadBufferSize is the default option for the maximum header size in KB for Bhojpur Application runtime HTTP servers.
	DefaultReadBufferSize = 4
)

// Config holds the Bhojpur Application Runtime Engine configuration.
type Config struct {
	ID                       string
	HTTPPort                 int
	PublicPort               *int
	ProfilePort              int
	EnableProfiling          bool
	APIGRPCPort              int
	InternalGRPCPort         int
	ApplicationPort          int
	APIListenAddresses       []string
	ApplicationProtocol      Protocol
	Mode                     utils.AppMode
	PlacementAddresses       []string
	GlobalConfig             string
	AllowedOrigins           string
	Standalone               config.StandaloneConfig
	Kubernetes               config.KubernetesConfig
	MaxConcurrency           int
	mtlsEnabled              bool
	SentryServiceAddress     string
	CertChain                *credentials.CertChain
	AppSSL                   bool
	MaxRequestBodySize       int
	UnixDomainSocket         string
	ReadBufferSize           int
	StreamRequestBody        bool
	GracefulShutdownDuration time.Duration
}

// NewRuntimeConfig returns a new runtime config.
func NewRuntimeConfig(
	id string, placementAddresses []string,
	controlPlaneAddress, allowedOrigins, globalConfig, componentsPath, appProtocol, mode string,
	httpPort, internalGRPCPort, apiGRPCPort int, apiListenAddresses []string, publicPort *int, appPort, profilePort int,
	enableProfiling bool, maxConcurrency int, mtlsEnabled bool, sentryAddress string, appSSL bool, maxRequestBodySize int, unixDomainSocket string, readBufferSize int, streamRequestBody bool, gracefulShutdownDuration time.Duration) *Config {
	return &Config{
		ID:                  id,
		HTTPPort:            httpPort,
		PublicPort:          publicPort,
		InternalGRPCPort:    internalGRPCPort,
		APIGRPCPort:         apiGRPCPort,
		ApplicationPort:     appPort,
		ProfilePort:         profilePort,
		APIListenAddresses:  apiListenAddresses,
		ApplicationProtocol: Protocol(appProtocol),
		Mode:                utils.AppMode(mode),
		PlacementAddresses:  placementAddresses,
		GlobalConfig:        globalConfig,
		AllowedOrigins:      allowedOrigins,
		Standalone: config.StandaloneConfig{
			ComponentsPath: componentsPath,
		},
		Kubernetes: config.KubernetesConfig{
			ControlPlaneAddress: controlPlaneAddress,
		},
		EnableProfiling:          enableProfiling,
		MaxConcurrency:           maxConcurrency,
		mtlsEnabled:              mtlsEnabled,
		SentryServiceAddress:     sentryAddress,
		AppSSL:                   appSSL,
		MaxRequestBodySize:       maxRequestBodySize,
		UnixDomainSocket:         unixDomainSocket,
		ReadBufferSize:           readBufferSize,
		StreamRequestBody:        streamRequestBody,
		GracefulShutdownDuration: gracefulShutdownDuration,
	}
}
