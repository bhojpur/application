package pubsub

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
	"github.com/google/uuid"

	svc_contenttype "github.com/bhojpur/service/pkg/contenttype"
	svc_pubsub "github.com/bhojpur/service/pkg/pubsub"
)

// CloudEvent is a request object to create a Bhojpur Application runtime compliant cloudevent.
type CloudEvent struct {
	ID              string
	Data            []byte
	Topic           string
	Pubsub          string
	DataContentType string
	TraceID         string
	TraceState      string
}

// NewCloudEvent encapsulates the creation of a Bhojpur Application cloudevent from an
// existing cloudevent or a raw payload.
func NewCloudEvent(req *CloudEvent) (map[string]interface{}, error) {
	if svc_contenttype.IsCloudEventContentType(req.DataContentType) {
		return svc_pubsub.FromCloudEvent(req.Data, req.Topic, req.Pubsub, req.TraceID, req.TraceState)
	}
	return svc_pubsub.NewCloudEventsEnvelope(uuid.New().String(), req.ID, svc_pubsub.DefaultCloudEventType,
		"", req.Topic, req.Pubsub, req.DataContentType, req.Data, req.TraceID, req.TraceState), nil
}
