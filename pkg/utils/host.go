package utils

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
	"net"
	"os"

	"github.com/pkg/errors"
)

const (
	// HostIPEnvVar is the environment variable to override host's chosen IP address.
	HostIPEnvVar = "APP_HOST_IP"
)

// GetHostAddress selects a valid outbound IP address for the host.
func GetHostAddress() (string, error) {
	if val, ok := os.LookupEnv(HostIPEnvVar); ok && val != "" {
		return val, nil
	}

	// Use UDP so no handshake is made.
	// Any IP can be used, since connection is not established, but we used a known DNS IP.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		// Could not find one via a UDP connection, so we fallback to the "old" way: try first non-loopback IPv4:
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return "", errors.Wrap(err, "error getting interface IP addresses")
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}

		return "", errors.New("could not determine host IP address")
	}

	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}
