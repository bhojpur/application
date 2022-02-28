package testing

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
	"testing"
	"time"
)

// WaitForListeningAddress waits for `addresses` to be listening for up to `timeout`.
func WaitForListeningAddress(t *testing.T, timeout time.Duration, addresses ...string) {
	start := time.Now().UTC()

	for _, address := range addresses {
		t.Logf("Waiting for address %q", address)
	check:
		for {
			d := timeout - time.Since(start)
			if d <= 0 {
				t.Log("Waiting for addresses timed out")
				t.FailNow()
			}

			conn, _ := net.DialTimeout("tcp", address, d)
			if conn != nil {
				conn.Close()

				break check
			}

			time.Sleep(time.Millisecond * 500)
		}
		t.Logf("Address %q is ready", address)
	}
}

// GetFreePorts asks the kernel for `num` free open ports that are ready to use.
// This code is retrofitted from freeport.GetFreePort().
func GetFreePorts(num uint) ([]int, error) {
	ports := make([]int, num)
	for i := uint(0); i < num; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		defer l.Close()

		ports[i] = l.Addr().(*net.TCPAddr).Port
	}

	return ports, nil
}
