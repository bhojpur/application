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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/bhojpur/application/pkg/api"
	"github.com/bhojpur/application/pkg/utils"
)

// Publish publishes payload to topic in pubsub referenced by pubsubName.
func (s *Standalone) Publish(publishAppID, pubsubName, topic string, payload []byte, socket string) error {
	if publishAppID == "" {
		return errors.New("publishAppID is missing")
	}

	if pubsubName == "" {
		return errors.New("pubsubName is missing")
	}

	if topic == "" {
		return errors.New("topic is missing")
	}

	l, err := s.process.List()
	if err != nil {
		return err
	}

	instance, err := getAppInstance(l, publishAppID)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://unix/v%s/publish/%s/%s", api.RuntimeAPIVersion, pubsubName, topic)

	var httpc http.Client
	if socket != "" {
		httpc.Transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", utils.GetSocket(socket, publishAppID, "http"))
			},
		}
	} else {
		url = fmt.Sprintf("http://localhost:%s/v%s/publish/%s/%s", fmt.Sprintf("%v", instance.HTTPPort), api.RuntimeAPIVersion, pubsubName, topic)
	}

	contentType := "application/json"

	// Detect publishing with CloudEvents envelope.
	var cloudEvent map[string]interface{}
	if json.Unmarshal(payload, &cloudEvent); err == nil {
		_, hasID := cloudEvent["id"]
		_, hasSource := cloudEvent["source"]
		_, hasSpecVersion := cloudEvent["specversion"]
		_, hasType := cloudEvent["type"]
		_, hasData := cloudEvent["data"]
		if hasID && hasSource && hasSpecVersion && hasType && hasData {
			contentType = "application/cloudevents+json"
		}
	}

	r, err := httpc.Post(url, contentType, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode >= 300 || r.StatusCode < 200 {
		return fmt.Errorf("unexpected status code %d on publishing to %s in %s", r.StatusCode, topic, pubsubName)
	}

	return nil
}

func getAppInstance(list []ListOutput, publishAppID string) (ListOutput, error) {
	for i := 0; i < len(list); i++ {
		if list[i].AppID == publishAppID {
			return list[i], nil
		}
	}
	return ListOutput{}, errors.New("couldn't find a running Bhojpur Application instance")
}
