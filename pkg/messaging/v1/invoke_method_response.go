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
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/anypb"

	commonv1pb "github.com/bhojpur/application/pkg/api/v1/common"
	internalv1pb "github.com/bhojpur/application/pkg/api/v1/internals"
)

// InvokeMethodResponse holds InternalInvokeResponse protobuf message
// and provides the helpers to manage it.
type InvokeMethodResponse struct {
	r *internalv1pb.InternalInvokeResponse
}

// NewInvokeMethodResponse returns new InvokeMethodResponse object with status.
func NewInvokeMethodResponse(statusCode int32, statusMessage string, statusDetails []*anypb.Any) *InvokeMethodResponse {
	return &InvokeMethodResponse{
		r: &internalv1pb.InternalInvokeResponse{
			Status:  &internalv1pb.Status{Code: statusCode, Message: statusMessage, Details: statusDetails},
			Message: &commonv1pb.InvokeResponse{},
		},
	}
}

// InternalInvokeResponse returns InvokeMethodResponse for InternalInvokeResponse pb to use the helpers.
func InternalInvokeResponse(resp *internalv1pb.InternalInvokeResponse) (*InvokeMethodResponse, error) {
	rsp := &InvokeMethodResponse{r: resp}
	if resp.Message == nil {
		resp.Message = &commonv1pb.InvokeResponse{Data: &anypb.Any{Value: []byte{}}}
	}

	return rsp, nil
}

// WithMessage sets InvokeResponse pb object to Message field.
func (imr *InvokeMethodResponse) WithMessage(pb *commonv1pb.InvokeResponse) *InvokeMethodResponse {
	imr.r.Message = pb
	return imr
}

// WithRawData sets Message using byte data and content type.
func (imr *InvokeMethodResponse) WithRawData(data []byte, contentType string) *InvokeMethodResponse {
	if contentType == "" {
		contentType = JSONContentType
	}

	imr.r.Message.ContentType = contentType

	// Clone data to prevent GC from deallocating data
	imr.r.Message.Data = &anypb.Any{Value: cloneBytes(data)}

	return imr
}

// WithHeaders sets gRPC response header metadata.
func (imr *InvokeMethodResponse) WithHeaders(headers metadata.MD) *InvokeMethodResponse {
	imr.r.Headers = MetadataToInternalMetadata(headers)
	return imr
}

// WithFastHTTPHeaders populates fasthttp response header to gRPC header metadata.
func (imr *InvokeMethodResponse) WithFastHTTPHeaders(header *fasthttp.ResponseHeader) *InvokeMethodResponse {
	md := AppInternalMetadata{}
	header.VisitAll(func(key []byte, value []byte) {
		md[string(key)] = &internalv1pb.ListStringValue{
			Values: []string{string(value)},
		}
	})
	if len(md) > 0 {
		imr.r.Headers = md
	}
	return imr
}

// WithTrailers sets Trailer in internal InvokeMethodResponse.
func (imr *InvokeMethodResponse) WithTrailers(trailer metadata.MD) *InvokeMethodResponse {
	imr.r.Trailers = MetadataToInternalMetadata(trailer)
	return imr
}

// Status gets Response status.
func (imr *InvokeMethodResponse) Status() *internalv1pb.Status {
	return imr.r.GetStatus()
}

// IsHTTPResponse returns true if response status code is http response status.
func (imr *InvokeMethodResponse) IsHTTPResponse() bool {
	// gRPC status code <= 15 - https://github.com/grpc/grpc/blob/master/doc/statuscodes.md
	// HTTP status code >= 100 - https://tools.ietf.org/html/rfc2616#section-10
	return imr.r.GetStatus().Code >= 100
}

// Proto clones the internal InvokeMethodResponse pb object.
func (imr *InvokeMethodResponse) Proto() *internalv1pb.InternalInvokeResponse {
	return imr.r
}

// Headers gets Headers metadata.
func (imr *InvokeMethodResponse) Headers() AppInternalMetadata {
	return imr.r.Headers
}

// Trailers gets Trailers metadata.
func (imr *InvokeMethodResponse) Trailers() AppInternalMetadata {
	return imr.r.Trailers
}

// Message returns message field in InvokeMethodResponse.
func (imr *InvokeMethodResponse) Message() *commonv1pb.InvokeResponse {
	return imr.r.Message
}

// RawData returns content_type and byte array body.
func (imr *InvokeMethodResponse) RawData() (string, []byte) {
	m := imr.r.Message
	if m == nil || m.GetData() == nil {
		return "", nil
	}

	contentType := m.GetContentType()
	dataTypeURL := m.GetData().GetTypeUrl()
	dataValue := m.GetData().GetValue()

	// set content_type to application/json only if typeurl is unset and data is given
	if contentType == "" && (dataTypeURL == "" && dataValue != nil) {
		contentType = JSONContentType
	}

	if dataTypeURL != "" {
		contentType = ProtobufContentType
	}

	return contentType, dataValue
}
