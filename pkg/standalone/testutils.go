package standalone

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
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/bhojpur/application/pkg/utils"
)

const SocketFormat = "/tmp/app-%s-http.socket"

type mockAppProcess struct {
	Lo  []ListOutput
	Err error
}

func (m *mockAppProcess) List() ([]ListOutput, error) {
	return m.Lo, m.Err
}

func getTestServerFunc(handler http.Handler) (*httptest.Server, int) {
	ts := httptest.NewUnstartedServer(handler)

	return ts, ts.Listener.Addr().(*net.TCPAddr).Port
}

func getTestServer(expectedPath, resp string) (*httptest.Server, int) {
	ts := httptest.NewUnstartedServer(handlerTestPathResp(expectedPath, resp))

	return ts, ts.Listener.Addr().(*net.TCPAddr).Port
}

func getTestSocketServerFunc(handler http.Handler, appID, path string) (*http.Server, net.Listener) {
	s := &http.Server{
		Handler: handler,
	}

	socket := utils.GetSocket(path, appID, "http")
	l, err := net.Listen("unix", socket)
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen on %v: %v", socket, err))
	}
	return s, l
}

func getTestSocketServer(expectedPath, resp, appID, path string) (*http.Server, net.Listener) {
	s := &http.Server{
		Handler: handlerTestPathResp(expectedPath, resp),
	}

	socket := utils.GetSocket(path, appID, "http")
	l, err := net.Listen("unix", socket)
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen on %v: %v", socket, err))
	}
	return s, l
}

func handlerTestPathResp(expectedPath, resp string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if expectedPath != "" && r.RequestURI != expectedPath {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
		if r.Method == http.MethodGet {
			w.Write([]byte(resp))
		} else {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)
			w.Write(buf.Bytes())
		}
	}
}
