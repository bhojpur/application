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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	k8s "k8s.io/client-go/kubernetes"

	scheme "github.com/bhojpur/application/pkg/client/clientset/versioned"

	//  azure auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"

	//  gcp auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	//  oidc auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	//  openstack auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	doOnce     sync.Once
	kubeconfig *string
)

func init() {
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

func getConfig() (*rest.Config, error) {
	doOnce.Do(func() {
		flag.Parse()
	})
	kubeConfigEnv := os.Getenv("KUBECONFIG")
	kubeConfigDelimiter := ":"
	if runtime.GOOS == "windows" {
		kubeConfigDelimiter = ";"
	}
	delimiterBelongsToPath := strings.Count(*kubeconfig, kubeConfigDelimiter) == 1 && strings.EqualFold(*kubeconfig, kubeConfigEnv)

	if len(kubeConfigEnv) != 0 && !delimiterBelongsToPath {
		kubeConfigs := strings.Split(kubeConfigEnv, kubeConfigDelimiter)
		if len(kubeConfigs) > 1 {
			return nil, fmt.Errorf("multiple kubeconfigs in KUBECONFIG environment variable - %s", kubeConfigEnv)
		}
		kubeconfig = &kubeConfigs[0]
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetKubeConfigClient returns the kubeconfig and the client created from the kubeconfig.
func GetKubeConfigClient() (*rest.Config, *k8s.Clientset, error) {
	config, err := getConfig()
	if err != nil {
		return nil, nil, err
	}
	client, err := k8s.NewForConfig(config)
	if err != nil {
		return config, nil, err
	}
	return config, client, nil
}

// Client returns a new Kubernetes client.
func Client() (*k8s.Clientset, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	return k8s.NewForConfig(config)
}

// AppClient returns a new Kubernetes Bhojpur Application client.
func AppClient() (scheme.Interface, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}
	return scheme.NewForConfig(config)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
