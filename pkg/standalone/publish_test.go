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
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bhojpur/application/pkg/utils"
)

func TestPublish(t *testing.T) {
	testCases := []struct {
		name          string
		publishAppID  string
		pubsubName    string
		payload       []byte
		topic         string
		lo            ListOutput
		listErr       error
		postResponse  string
		handler       http.HandlerFunc
		errorExpected bool
		errString     string
	}{
		{
			name:          "test empty appID",
			publishAppID:  "",
			payload:       []byte("test"),
			pubsubName:    "test",
			errString:     "publishAppID is missing",
			errorExpected: true,
			handler:       handlerTestPathResp("", ""),
		},
		{
			name:          "test empty topic",
			publishAppID:  "test",
			payload:       []byte("test"),
			pubsubName:    "test",
			errString:     "topic is missing",
			errorExpected: true,
			handler:       handlerTestPathResp("", ""),
		},
		{
			name:          "test empty pubsubName",
			publishAppID:  "test",
			payload:       []byte("test"),
			topic:         "test",
			errString:     "pubsubName is missing",
			errorExpected: true,
			handler:       handlerTestPathResp("", ""),
		},
		{
			name:          "test list error",
			publishAppID:  "test",
			payload:       []byte("test"),
			topic:         "test",
			pubsubName:    "test",
			listErr:       assert.AnError,
			errString:     assert.AnError.Error(),
			errorExpected: true,
			handler:       handlerTestPathResp("", ""),
		},
		{
			name:         "test empty appID in list output",
			publishAppID: "test",
			payload:      []byte("test"),
			topic:        "test",
			pubsubName:   "test",
			lo: ListOutput{
				// empty appID
				Command: "test",
			},
			errString:     "couldn't find a running Bhojpur Application instance",
			errorExpected: true,
			handler:       handlerTestPathResp("", ""),
		},
		{
			name:         "successful call not found",
			publishAppID: "myAppID",
			pubsubName:   "testPubsubName",
			topic:        "testTopic",
			payload:      []byte("test payload"),
			lo: ListOutput{
				AppID: "not my myAppID",
			},
			errString:     "couldn't find a running Bhojpur Application instance",
			errorExpected: true,
			handler:       handlerTestPathResp("", ""),
		},
		{
			name:         "successful call",
			publishAppID: "myAppID",
			pubsubName:   "testPubsubName",
			topic:        "testTopic",
			payload:      []byte("test payload"),
			postResponse: "test payload",
			lo: ListOutput{
				AppID: "myAppID",
			},
			handler: handlerTestPathResp("/v1/publish/testPubsubName/testTopic", ""),
		},
		{
			name:         "successful cloudevent envelope",
			publishAppID: "myAppID",
			pubsubName:   "testPubsubName",
			topic:        "testTopic",
			payload:      []byte(`{"id": "1234", "source": "test", "specversion": "1.0", "type": "product.v1", "datacontenttype": "application/json", "data": {"id": "test", "description": "Testing 12345"}}`),
			postResponse: "test payload",
			lo: ListOutput{
				AppID: "myAppID",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Content-Type") != "application/cloudevents+json" {
					w.WriteHeader(http.StatusInternalServerError)

					return
				}
				if r.Method == http.MethodGet {
					w.Write([]byte(""))
				} else {
					buf := new(bytes.Buffer)
					buf.ReadFrom(r.Body)
					w.Write(buf.Bytes())
				}
			},
		},
	}
	for _, socket := range []string{"", "/tmp"} {
		// TODO(@daixiang0): add Windows support
		if runtime.GOOS == "windows" && socket != "" {
			continue
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if socket != "" {
					ts, l := getTestSocketServerFunc(tc.handler, tc.publishAppID, socket)
					go ts.Serve(l)
					defer func() {
						l.Close()
						for _, protocol := range []string{"http", "grpc"} {
							os.Remove(utils.GetSocket(socket, tc.publishAppID, protocol))
						}
					}()
				} else {
					ts, port := getTestServerFunc(tc.handler)
					ts.Start()
					defer ts.Close()
					tc.lo.HTTPPort = port
				}

				client := &Standalone{
					process: &mockAppProcess{
						Lo:  []ListOutput{tc.lo},
						Err: tc.listErr,
					},
				}
				err := client.Publish(tc.publishAppID, tc.pubsubName, tc.topic, tc.payload, socket)
				if tc.errorExpected {
					assert.Error(t, err, "expected an error")
					assert.Equal(t, tc.errString, err.Error(), "expected error strings to match")
				} else {
					assert.NoError(t, err, "expected no error")
				}
			})
		}
	}
}
