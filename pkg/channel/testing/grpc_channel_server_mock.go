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
	context "context"
	"encoding/json"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	commonv1pb "github.com/bhojpur/api/pkg/core/v1/common"
	runtimev1pb "github.com/bhojpur/api/pkg/core/v1/runtime"
)

// MockServer implementation of fake user Bhojpur Application server.
type MockServer struct {
	Error                    error
	Subscriptions            []*runtimev1pb.TopicSubscription
	Bindings                 []string
	BindingEventResponse     runtimev1pb.BindingEventResponse
	TopicEventResponseStatus runtimev1pb.TopicEventResponse_TopicEventResponseStatus
}

func (m *MockServer) OnInvoke(ctx context.Context, in *commonv1pb.InvokeRequest) (*commonv1pb.InvokeResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	dt := map[string]string{
		"method": in.Method,
	}

	for k, v := range md {
		dt[k] = v[0]
	}

	dt["httpverb"] = in.HttpExtension.GetVerb().String()
	dt["querystring"] = in.HttpExtension.Querystring

	ds, _ := json.Marshal(dt)
	return &commonv1pb.InvokeResponse{Data: &anypb.Any{Value: ds}, ContentType: "application/json"}, m.Error
}

func (m *MockServer) ListTopicSubscriptions(ctx context.Context, in *emptypb.Empty) (*runtimev1pb.ListTopicSubscriptionsResponse, error) {
	return &runtimev1pb.ListTopicSubscriptionsResponse{
		Subscriptions: m.Subscriptions,
	}, m.Error
}

func (m *MockServer) ListInputBindings(ctx context.Context, in *emptypb.Empty) (*runtimev1pb.ListInputBindingsResponse, error) {
	return &runtimev1pb.ListInputBindingsResponse{
		Bindings: m.Bindings,
	}, m.Error
}

func (m *MockServer) OnBindingEvent(ctx context.Context, in *runtimev1pb.BindingEventRequest) (*runtimev1pb.BindingEventResponse, error) {
	return &m.BindingEventResponse, m.Error
}

func (m *MockServer) OnTopicEvent(ctx context.Context, in *runtimev1pb.TopicEventRequest) (*runtimev1pb.TopicEventResponse, error) {
	return &runtimev1pb.TopicEventResponse{
		Status: m.TopicEventResponseStatus,
	}, m.Error
}
