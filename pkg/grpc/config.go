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

// ServerConfig is the config object for a Bhojpur Application runtime gRPC server.
type ServerConfig struct {
	AppID              string
	HostAddress        string
	Port               int
	APIListenAddresses []string
	NameSpace          string
	TrustDomain        string
	MaxRequestBodySize int
	UnixDomainSocket   string
	ReadBufferSize     int
}

// NewServerConfig returns a new Bhojpur Application runtime gRPC server config.
func NewServerConfig(appID string, hostAddress string, port int, apiListenAddresses []string, namespace string, trustDomain string, maxRequestBodySize int, unixDomainSocket string, readBufferSize int) ServerConfig {
	return ServerConfig{
		AppID:              appID,
		HostAddress:        hostAddress,
		Port:               port,
		APIListenAddresses: apiListenAddresses,
		NameSpace:          namespace,
		TrustDomain:        trustDomain,
		MaxRequestBodySize: maxRequestBodySize,
		UnixDomainSocket:   unixDomainSocket,
		ReadBufferSize:     readBufferSize,
	}
}
