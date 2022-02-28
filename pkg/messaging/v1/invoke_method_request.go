package v1

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
	"errors"
	"strings"

	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/types/known/anypb"

	commonv1pb "github.com/bhojpur/application/pkg/api/v1/common"
	internalv1pb "github.com/bhojpur/application/pkg/api/v1/internals"
)

const (
	// DefaultAPIVersion is the default Bhojpur Application runtime API version.
	DefaultAPIVersion = internalv1pb.APIVersion_V1
)

// InvokeMethodRequest holds InternalInvokeRequest protobuf message
// and provides the helpers to manage it.
type InvokeMethodRequest struct {
	r *internalv1pb.InternalInvokeRequest
}

// NewInvokeMethodRequest creates InvokeMethodRequest object for method.
func NewInvokeMethodRequest(method string) *InvokeMethodRequest {
	return &InvokeMethodRequest{
		r: &internalv1pb.InternalInvokeRequest{
			Ver: DefaultAPIVersion,
			Message: &commonv1pb.InvokeRequest{
				Method: method,
			},
		},
	}
}

// FromInvokeRequestMessage creates InvokeMethodRequest object from InvokeRequest pb object.
func FromInvokeRequestMessage(pb *commonv1pb.InvokeRequest) *InvokeMethodRequest {
	return &InvokeMethodRequest{
		r: &internalv1pb.InternalInvokeRequest{
			Ver:     DefaultAPIVersion,
			Message: pb,
		},
	}
}

// InternalInvokeRequest creates InvokeMethodRequest object from InternalInvokeRequest pb object.
func InternalInvokeRequest(pb *internalv1pb.InternalInvokeRequest) (*InvokeMethodRequest, error) {
	req := &InvokeMethodRequest{r: pb}
	if pb.Message == nil {
		return nil, errors.New("Message field is nil")
	}

	return req, nil
}

// WithActor sets actor type and id.
func (imr *InvokeMethodRequest) WithActor(actorType, actorID string) *InvokeMethodRequest {
	imr.r.Actor = &internalv1pb.Actor{ActorType: actorType, ActorId: actorID}
	return imr
}

// WithMetadata sets metadata.
func (imr *InvokeMethodRequest) WithMetadata(md map[string][]string) *InvokeMethodRequest {
	imr.r.Metadata = MetadataToInternalMetadata(md)
	return imr
}

// WithFastHTTPHeaders sets fasthttp request headers.
func (imr *InvokeMethodRequest) WithFastHTTPHeaders(header *fasthttp.RequestHeader) *InvokeMethodRequest {
	md := map[string][]string{}
	header.VisitAll(func(key []byte, value []byte) {
		md[string(key)] = []string{string(value)}
	})
	imr.r.Metadata = MetadataToInternalMetadata(md)
	return imr
}

// WithRawData sets message data and content_type.
func (imr *InvokeMethodRequest) WithRawData(data []byte, contentType string) *InvokeMethodRequest {
	if contentType == "" {
		contentType = JSONContentType
	}
	imr.r.Message.ContentType = contentType
	imr.r.Message.Data = &anypb.Any{Value: data}
	return imr
}

// WithHTTPExtension sets new HTTP extension with verb and querystring.
func (imr *InvokeMethodRequest) WithHTTPExtension(verb string, querystring string) *InvokeMethodRequest {
	httpMethod, ok := commonv1pb.HTTPExtension_Verb_value[strings.ToUpper(verb)]
	if !ok {
		httpMethod = int32(commonv1pb.HTTPExtension_POST)
	}

	imr.r.Message.HttpExtension = &commonv1pb.HTTPExtension{
		Verb:        commonv1pb.HTTPExtension_Verb(httpMethod),
		Querystring: querystring,
	}

	return imr
}

// WithCustomHTTPMetadata applies a metadata map to a InvokeMethodRequest.
func (imr *InvokeMethodRequest) WithCustomHTTPMetadata(md map[string]string) *InvokeMethodRequest {
	for k, v := range md {
		if imr.r.Metadata == nil {
			imr.r.Metadata = make(map[string]*internalv1pb.ListStringValue)
		}

		// NOTE: We don't explicitly lowercase the keys here but this will be done
		//       later when attached to the HTTP request as headers.
		imr.r.Metadata[k] = &internalv1pb.ListStringValue{Values: []string{v}}
	}

	return imr
}

// EncodeHTTPQueryString generates querystring for http using http extension object.
func (imr *InvokeMethodRequest) EncodeHTTPQueryString() string {
	m := imr.r.Message
	if m == nil || m.GetHttpExtension() == nil {
		return ""
	}

	return m.GetHttpExtension().Querystring
}

// APIVersion gets API version of InvokeMethodRequest.
func (imr *InvokeMethodRequest) APIVersion() internalv1pb.APIVersion {
	return imr.r.GetVer()
}

// Metadata gets Metadata of InvokeMethodRequest.
func (imr *InvokeMethodRequest) Metadata() AppInternalMetadata {
	return imr.r.GetMetadata()
}

// Proto returns InternalInvokeRequest Proto object.
func (imr *InvokeMethodRequest) Proto() *internalv1pb.InternalInvokeRequest {
	return imr.r
}

// Actor returns actor type and id.
func (imr *InvokeMethodRequest) Actor() *internalv1pb.Actor {
	return imr.r.GetActor()
}

// Message gets InvokeRequest Message object.
func (imr *InvokeMethodRequest) Message() *commonv1pb.InvokeRequest {
	return imr.r.Message
}

// RawData returns content_type and byte array body.
func (imr *InvokeMethodRequest) RawData() (string, []byte) {
	m := imr.r.Message
	if m == nil || m.Data == nil {
		return "", nil
	}

	contentType := m.GetContentType()
	dataTypeURL := m.GetData().GetTypeUrl()
	dataValue := m.GetData().GetValue()

	// set content_type to application/json only if typeurl is unset and data is given
	if contentType == "" && (dataTypeURL == "" && dataValue != nil) {
		contentType = JSONContentType
	}

	return contentType, dataValue
}

// Adds a new header to the existing set.
func (imr *InvokeMethodRequest) AddHeaders(header *fasthttp.RequestHeader) {
	md := map[string][]string{}
	header.VisitAll(func(key []byte, value []byte) {
		md[string(key)] = []string{string(value)}
	})

	internalMd := MetadataToInternalMetadata(md)

	if imr.r.Metadata == nil {
		imr.r.Metadata = internalMd
	} else {
		for key, val := range internalMd {
			// We're only adding new values, not overwriting existing
			if _, ok := imr.r.Metadata[key]; !ok {
				imr.r.Metadata[key] = val
			}
		}
	}
}
