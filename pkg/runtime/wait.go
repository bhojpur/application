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
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	timeoutSeconds       int    = 60
	requestTimeoutMillis int    = 500
	periodMillis         int    = 100
	urlFormat            string = "http://localhost:%s/v1.0/healthz/outbound"
)

func waitUntilAppOutboundReady(appHTTPPort string) {
	outboundReadyHealthURL := fmt.Sprintf(urlFormat, appHTTPPort)
	client := &http.Client{
		Timeout: time.Duration(requestTimeoutMillis) * time.Millisecond,
	}
	println(fmt.Sprintf("Waiting for Bhojpur Application runtime to be outbound ready (timeout: %d seconds): url=%s\n", timeoutSeconds, outboundReadyHealthURL))

	var err error
	timeoutAt := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	lastPrintErrorTime := time.Now()
	for time.Now().Before(timeoutAt) {
		err = checkIfOutboundReady(client, outboundReadyHealthURL)
		if err == nil {
			println("Bhojpur Application runtime is outbound ready!")
			return
		}

		if time.Now().After(lastPrintErrorTime) {
			// print the error once in one seconds to avoid too many errors
			lastPrintErrorTime = time.Now().Add(time.Second)
			println(fmt.Sprintf("Bhojpur Application runtime outbound NOT ready yet: %v", err))
		}

		time.Sleep(time.Duration(periodMillis) * time.Millisecond)
	}

	println(fmt.Sprintf("timeout waiting for Bhojpur Application runtime to become outbound ready. Last error: %v", err))
}

func checkIfOutboundReady(client *http.Client, outboundReadyHealthURL string) error {
	req, err := http.NewRequest(http.MethodGet, outboundReadyHealthURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("HTTP status code %v", resp.StatusCode)
	}

	return nil
}
