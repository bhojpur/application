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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	core_v1 "k8s.io/api/core/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// PortForward provides a port-forward connection in a kubernetes cluster.
type PortForward struct {
	Config     *rest.Config
	Method     string
	URL        *url.URL
	Host       string
	LocalPort  int
	RemotePort int
	EmitLogs   bool
	StopCh     chan struct{}
	ReadyCh    chan struct{}
}

// NewPortForward returns an instance of PortForward struct that can be used
// for establishing port-forwarding connection to a pod in kubernetes cluster,
// specified by namespace and deployName.
func NewPortForward(
	config *rest.Config,
	namespace, deployName string,
	host string, localPort, remotePort int,
	emitLogs bool,
) (*PortForward, error) {
	client, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	podList, err := ListPods(client, namespace, nil)
	if err != nil {
		return nil, err
	}

	podName := ""

	for _, pod := range podList.Items {
		if pod.Status.Phase == core_v1.PodRunning {
			if strings.HasPrefix(pod.Name, deployName) {
				podName = pod.Name
				break
			}
		}
	}

	if podName == "" {
		return nil, fmt.Errorf("no running pods found for %s", deployName)
	}

	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	return &PortForward{
		Config:     config,
		Method:     "POST",
		URL:        req.URL(),
		Host:       host,
		LocalPort:  localPort,
		RemotePort: remotePort,
		EmitLogs:   emitLogs,
		StopCh:     make(chan struct{}, 1),
		ReadyCh:    make(chan struct{}),
	}, nil
}

// run creates port-forward connection and blocks
// until Stop() is called.
func (pf *PortForward) run() error {
	transport, upgrader, err := spdy.RoundTripperFor(pf.Config)
	if err != nil {
		return err
	}

	out := ioutil.Discard
	errOut := ioutil.Discard
	if pf.EmitLogs {
		out = os.Stdout
		errOut = os.Stderr
	}

	ports := []string{fmt.Sprintf("%d:%d", pf.LocalPort, pf.RemotePort)}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, pf.Method, pf.URL)

	fw, err := portforward.NewOnAddresses(dialer, []string{pf.Host}, ports, pf.StopCh, pf.ReadyCh, out, errOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}

// Init creates and runs a port-forward connection.
// This function blocks until connection is established.
// Note: Caller should call Stop() to finish the connection.
func (pf *PortForward) Init() error {
	failure := make(chan error)

	go func() {
		if err := pf.run(); err != nil {
			failure <- err
		}
	}()

	select {
	// if `pf.run()` succeeds, block until terminated
	case <-pf.ReadyCh:

	// if failure, causing a receive `<-failure` and returns the error
	case err := <-failure:
		return err
	}

	return nil
}

// Stop terminates port-forwarding connection.
func (pf *PortForward) Stop() {
	close(pf.StopCh)
}

// GetStop returns StopCh for a PortForward instance.
// Receiving on StopCh will block until the port forwarding stops.
func (pf *PortForward) GetStop() <-chan struct{} {
	return pf.StopCh
}
