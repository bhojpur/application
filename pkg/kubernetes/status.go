package kubernetes

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
	"fmt"
	"os"
	"strings"
	"sync"

	k8s "k8s.io/client-go/kubernetes"

	"github.com/bhojpur/application/pkg/utils"
)

var controlPlaneLabels = []string{
	"app-operator",
	"app-sentry",
	"app-placement", // TODO: This is for the backward compatibility.
	"app-placement-server",
	"app-sidecar-injector",
	"app-dashboard",
}

type StatusClient struct {
	client k8s.Interface
}

// StatusOutput represents the status of a named Bhojpur Application resource.
type StatusOutput struct {
	Name      string `csv:"NAME"`
	Namespace string `csv:"NAMESPACE"`
	Healthy   string `csv:"HEALTHY"`
	Status    string `csv:"STATUS"`
	Replicas  int    `csv:"REPLICAS"`
	Version   string `csv:"VERSION"`
	Age       string `csv:"AGE"`
	Created   string `csv:"CREATED"`
}

// Create a new k8s client for status commands.
func NewStatusClient() (*StatusClient, error) {
	clientset, err := Client()
	if err != nil {
		return nil, err
	}
	return &StatusClient{
		client: clientset,
	}, nil
}

// List status for Bhojpur Application resources.
func (s *StatusClient) Status() ([]StatusOutput, error) {
	client := s.client
	if client == nil {
		return nil, errors.New("kubernetes client not initialized")
	}
	var wg sync.WaitGroup
	wg.Add(len(controlPlaneLabels))

	m := sync.Mutex{}
	statuses := []StatusOutput{}

	for _, lbl := range controlPlaneLabels {
		go func(label string) {
			defer wg.Done()
			// Query all namespaces for Bhojpur Application pods.
			p, err := ListPodsInterface(client, map[string]string{
				"app": label,
			})
			if err != nil {
				utils.WarningStatusEvent(os.Stdout, "Failed to get status for %s: %s", label, err.Error())
				return
			}

			if len(p.Items) == 0 {
				return
			}
			pod := p.Items[0]
			replicas := len(p.Items)
			image := pod.Spec.Containers[0].Image
			namespace := pod.GetNamespace()
			age := utils.GetAge(pod.CreationTimestamp.Time)
			created := pod.CreationTimestamp.Format("2006-01-02 15:04.05")
			version := image[strings.IndexAny(image, ":")+1:]
			status := ""

			// loop through all replicas and update to Running/Healthy status only if all instances are Running and Healthy
			healthy := "False"
			running := true

			for _, p := range p.Items {
				if len(p.Status.ContainerStatuses) == 0 {
					status = string(p.Status.Phase)
				} else if p.Status.ContainerStatuses[0].State.Waiting != nil {
					status = fmt.Sprintf("Waiting (%s)", p.Status.ContainerStatuses[0].State.Waiting.Reason)
				} else if pod.Status.ContainerStatuses[0].State.Terminated != nil {
					status = "Terminated"
				}

				if len(p.Status.ContainerStatuses) == 0 ||
					p.Status.ContainerStatuses[0].State.Running == nil {
					running = false

					break
				}

				if p.Status.ContainerStatuses[0].Ready {
					healthy = "True"
				}
			}

			if running {
				status = "Running"
			}

			s := StatusOutput{
				Name:      label,
				Namespace: namespace,
				Created:   created,
				Age:       age,
				Status:    status,
				Version:   version,
				Healthy:   healthy,
				Replicas:  replicas,
			}

			m.Lock()
			statuses = append(statuses, s)
			m.Unlock()
		}(lbl)
	}

	wg.Wait()
	return statuses, nil
}
