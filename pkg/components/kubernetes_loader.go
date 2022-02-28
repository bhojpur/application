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
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"

	"github.com/bhojpur/service/pkg/utils/logger"

	operatorv1pb "github.com/bhojpur/application/pkg/api/v1/operator"
	config "github.com/bhojpur/application/pkg/config/modes"
	components_v1alpha1 "github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
)

var log = logger.NewLogger("app.runtime.components")

const (
	operatorCallTimeout = time.Second * 5
	operatorMaxRetries  = 100
)

// KubernetesComponents loads components in a kubernetes environment.
type KubernetesComponents struct {
	config    config.KubernetesConfig
	client    operatorv1pb.OperatorClient
	namespace string
	podName   string
}

// NewKubernetesComponents returns a new kubernetes loader.
func NewKubernetesComponents(configuration config.KubernetesConfig, namespace string, operatorClient operatorv1pb.OperatorClient, podName string) *KubernetesComponents {
	return &KubernetesComponents{
		config:    configuration,
		client:    operatorClient,
		namespace: namespace,
		podName:   podName,
	}
}

// LoadComponents returns components from a given control plane address.
func (k *KubernetesComponents) LoadComponents() ([]components_v1alpha1.Component, error) {
	resp, err := k.client.ListComponents(context.Background(), &operatorv1pb.ListComponentsRequest{
		Namespace: k.namespace,
		PodName:   k.podName,
	}, grpc_retry.WithMax(operatorMaxRetries), grpc_retry.WithPerRetryTimeout(operatorCallTimeout))
	if err != nil {
		return nil, err
	}
	comps := resp.GetComponents()

	components := []components_v1alpha1.Component{}
	for _, c := range comps {
		var component components_v1alpha1.Component
		component.Spec = components_v1alpha1.ComponentSpec{}
		err := json.Unmarshal(c, &component)
		if err != nil {
			log.Warnf("error deserializing component: %s", err)
			continue
		}
		components = append(components, component)
	}
	return components, nil
}
