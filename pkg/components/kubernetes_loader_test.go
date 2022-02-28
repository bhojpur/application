package components

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
	"net"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	operatorv1pb "github.com/bhojpur/application/pkg/api/v1/operator"
	config "github.com/bhojpur/application/pkg/config/modes"
	"github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
	subscriptions "github.com/bhojpur/application/pkg/kubernetes/subscriptions/v1alpha1"
)

type mockOperator struct {
	operatorv1pb.UnimplementedOperatorServer
}

func (o *mockOperator) GetConfiguration(ctx context.Context, in *operatorv1pb.GetConfigurationRequest) (*operatorv1pb.GetConfigurationResponse, error) {
	return nil, nil
}

func (o *mockOperator) ListComponents(ctx context.Context, in *operatorv1pb.ListComponentsRequest) (*operatorv1pb.ListComponentResponse, error) {
	component := v1alpha1.Component{}
	component.ObjectMeta.Name = "test"
	component.ObjectMeta.Labels = map[string]string{
		"podName": in.PodName,
	}
	component.Spec = v1alpha1.ComponentSpec{
		Type: "testtype",
	}
	b, _ := json.Marshal(&component)

	return &operatorv1pb.ListComponentResponse{
		Components: [][]byte{b},
	}, nil
}

func (o *mockOperator) ListSubscriptionsV2(ctx context.Context, in *operatorv1pb.ListSubscriptionsRequest) (*operatorv1pb.ListSubscriptionsResponse, error) {
	subscription := subscriptions.Subscription{}
	subscription.ObjectMeta.Name = "test"
	subscription.Spec = subscriptions.SubscriptionSpec{
		Topic:      "topic",
		Route:      "route",
		Pubsubname: "pubsub",
	}
	b, _ := json.Marshal(&subscription)

	return &operatorv1pb.ListSubscriptionsResponse{
		Subscriptions: [][]byte{b},
	}, nil
}

func (o *mockOperator) ComponentUpdate(in *operatorv1pb.ComponentUpdateRequest, srv operatorv1pb.Operator_ComponentUpdateServer) error {
	return nil
}

func getOperatorClient(address string) operatorv1pb.OperatorClient {
	conn, _ := grpc.Dial(address, grpc.WithInsecure())
	return operatorv1pb.NewOperatorClient(conn)
}

func TestLoadComponents(t *testing.T) {
	port, _ := freeport.GetFreePort()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	assert.NoError(t, err)

	s := grpc.NewServer()
	operatorv1pb.RegisterOperatorServer(s, &mockOperator{})
	defer s.Stop()

	go func() {
		s.Serve(lis)
	}()

	time.Sleep(time.Second * 1)

	request := &KubernetesComponents{
		client: getOperatorClient(fmt.Sprintf("localhost:%d", port)),
		config: config.KubernetesConfig{
			ControlPlaneAddress: fmt.Sprintf("localhost:%v", port),
		},
		podName: "testPodName",
	}

	response, err := request.LoadComponents()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test", response[0].Name)
	assert.Equal(t, "testtype", response[0].Spec.Type)
	assert.Equal(t, "testPodName", response[0].ObjectMeta.Labels["podName"])
}
