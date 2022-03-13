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
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	apisvr "github.com/bhojpur/api/pkg/core"
	"github.com/bhojpur/application/pkg/utils"
)

// Invoke is a command to invoke a remote or local Bhojpur Application instance.
func (s *Standalone) Invoke(appID, method string, data []byte, verb string, path string) (string, error) {
	list, err := s.process.List()
	if err != nil {
		return "", err
	}

	for _, lo := range list {
		if lo.AppID == appID {
			url := makeEndpoint(lo, method)
			req, err := http.NewRequest(verb, url, bytes.NewBuffer(data))
			if err != nil {
				return "", err
			}
			req.Header.Set("Content-Type", "application/json")

			var httpc http.Client

			if path != "" {
				httpc.Transport = &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", utils.GetSocket(path, appID, "http"))
					},
				}
			}

			r, err := httpc.Do(req)
			if err != nil {
				return "", err
			}
			defer r.Body.Close()
			return handleResponse(r)
		}
	}

	return "", fmt.Errorf("app ID %s not found", appID)
}

func makeEndpoint(lo ListOutput, method string) string {
	return fmt.Sprintf("http://127.0.0.1:%s/v%s/invoke/%s/method/%s", fmt.Sprintf("%v", lo.HTTPPort), apisvr.RuntimeAPIVersion, lo.AppID, method)
}

func handleResponse(response *http.Response) (string, error) {
	rb, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if len(rb) > 0 {
		return string(rb), nil
	}

	return "", nil
}
