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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	retryablehttp "github.com/hashicorp/go-retryablehttp"

	apisvr "github.com/bhojpur/api/pkg/core"
)

// Get retrieves the metadata of a given Bhojpur Application's sidecar.
func Get(httpPort int, appID, socket string) (*apisvr.Metadata, error) {
	url := makeMetadataGetEndpoint(httpPort)

	var httpc http.Client
	if socket != "" {
		fileInfo, err := os.Stat(socket)
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			socket = GetSocket(socket, appID, "http")
		}

		httpc.Transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		}
	}

	r, err := httpc.Get(url)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()
	return handleMetadataResponse(r)
}

// Put sets one metadata attribute on a given app's sidecar.
func Put(httpPort int, key, value, appID, socket string) error {
	client := retryablehttp.NewClient()
	client.Logger = nil

	if socket != "" {
		client.HTTPClient.Transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", GetSocket(socket, appID, "http"))
			},
		}
	}

	url := makeMetadataPutEndpoint(httpPort, key)

	req, err := retryablehttp.NewRequest("PUT", url, strings.NewReader(value))
	if err != nil {
		return err
	}

	r, err := client.Do(req)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	if socket != "" {
		// Retryablehttp does not close idle socket connections.
		defer client.HTTPClient.CloseIdleConnections()
	}
	return nil
}

func makeMetadataGetEndpoint(httpPort int) string {
	if httpPort == 0 {
		return fmt.Sprintf("http://unix/v%s/metadata", apisvr.RuntimeAPIVersion)
	}
	return fmt.Sprintf("http://127.0.0.1:%v/v%s/metadata", httpPort, apisvr.RuntimeAPIVersion)
}

func makeMetadataPutEndpoint(httpPort int, key string) string {
	if httpPort == 0 {
		return fmt.Sprintf("http://unix/v%s/metadata/%s", apisvr.RuntimeAPIVersion, key)
	}
	return fmt.Sprintf("http://127.0.0.1:%v/v%s/metadata/%s", httpPort, apisvr.RuntimeAPIVersion, key)
}

func handleMetadataResponse(response *http.Response) (*apisvr.Metadata, error) {
	rb, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var m apisvr.Metadata
	err = json.Unmarshal(rb, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
