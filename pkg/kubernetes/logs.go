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
	"context"
	"fmt"
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"
)

const (
	appContainerName      = "appsvr"
	appIDContainerArgName = "--app-id"
)

// Logs fetches Bhojpur Application sidecar logs from Kubernetes.
func Logs(appID, podName, namespace string) error {
	client, err := Client()
	if err != nil {
		return err
	}

	if namespace == "" {
		namespace = corev1.NamespaceDefault
	}

	pods, err := ListPods(client, namespace, nil)
	if err != nil {
		return fmt.Errorf("could not get Bhojpur Application runtime logs %v", err)
	}

	if podName == "" {
		// no pod name specified. in case of multiple pods, the first one will be selected
		var foundAppPod bool
		for _, pod := range pods.Items {
			if foundAppPod {
				break
			}
			for _, container := range pod.Spec.Containers {
				if container.Name == appContainerName {
					// find app ID
					for i, arg := range container.Args {
						if arg == appIDContainerArgName {
							id := container.Args[i+1]
							if id == appID {
								podName = pod.Name
								foundAppPod = true
								break
							}
						}
					}
				}
			}
		}
		if !foundAppPod {
			return fmt.Errorf("could not get logs. Please check app-id (%s) and namespace (%s)", appID, namespace)
		}
	}

	getLogsRequest := client.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Container: appContainerName, Follow: false})
	logStream, err := getLogsRequest.Stream(context.TODO())
	if err != nil {
		return fmt.Errorf("could not get logs. Please check pod-name (%s). Error - %v", podName, err)
	}
	defer logStream.Close()
	_, err = io.Copy(os.Stdout, logStream)
	if err != nil {
		return fmt.Errorf("could not get logs %v", err)
	}

	return nil
}
