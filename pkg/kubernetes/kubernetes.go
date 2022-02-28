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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	helm "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/helm/pkg/strvals"

	"github.com/bhojpur/application/pkg/utils"
	cli_ver "github.com/bhojpur/application/pkg/version"
)

const (
	appReleaseName = "bhojpur"
	appHelmRepo    = "https://repo.bhojpur.net/helm-charts"
	latestVersion  = "latest"
)

type InitConfiguration struct {
	Version    string
	Namespace  string
	EnableMTLS bool
	EnableHA   bool
	Args       []string
	Wait       bool
	Timeout    uint
}

// Init deploys the Bhojpur Application operator using the supplied runtime version.
func Init(config InitConfiguration) error {
	msg := "Deploying the Bhojpur Application control plane to your cluster..."

	stopSpinning := utils.Spinner(os.Stdout, msg)
	defer stopSpinning(utils.Failure)

	err := install(config)
	if err != nil {
		return err
	}

	stopSpinning(utils.Success)

	return nil
}

func createNamespace(namespace string) error {
	_, client, err := GetKubeConfigClient()
	if err != nil {
		return fmt.Errorf("can't connect to a Kubernetes cluster: %v", err)
	}

	ns := &v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: namespace,
		},
	}
	// try to create the namespace if it doesn't exist. ok to ignore error.
	client.CoreV1().Namespaces().Create(context.TODO(), ns, meta_v1.CreateOptions{})
	return nil
}

func helmConfig(namespace string) (*helm.Configuration, error) {
	ac := helm.Configuration{}
	flags := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}
	err := ac.Init(flags, namespace, "secret", debugLogf)
	return &ac, err
}

func getVersion(version string) (string, error) {
	if version == latestVersion {
		var err error
		version, err = cli_ver.GetAppVersion()
		if err != nil {
			return "", fmt.Errorf("cannot get the latest release version: %s", err)
		}
		version = strings.TrimPrefix(version, "v")
	}
	return version, nil
}

func createTempDir() (string, error) {
	dir, err := ioutil.TempDir("", "bhojpur")
	if err != nil {
		return "", fmt.Errorf("error creating temp dir: %s", err)
	}
	return dir, nil
}

func locateChartFile(dirPath string) (string, error) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(dirPath, files[0].Name()), nil
}

func appChart(version string, config *helm.Configuration) (*chart.Chart, error) {
	pull := helm.NewPull()
	pull.RepoURL = appHelmRepo
	pull.Settings = &cli.EnvSettings{}

	if version != latestVersion {
		pull.Version = chartVersion(version)
	}

	dir, err := createTempDir()
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	pull.DestDir = dir

	_, err = pull.Run(appReleaseName)
	if err != nil {
		return nil, err
	}

	chartPath, err := locateChartFile(dir)
	if err != nil {
		return nil, err
	}
	return loader.Load(chartPath)
}

func chartValues(config InitConfiguration) (map[string]interface{}, error) {
	chartVals := map[string]interface{}{}
	globalVals := []string{
		fmt.Sprintf("global.ha.enabled=%t", config.EnableHA),
		fmt.Sprintf("global.mtls.enabled=%t", config.EnableMTLS),
	}
	globalVals = append(globalVals, config.Args...)

	for _, v := range globalVals {
		if err := strvals.ParseInto(v, chartVals); err != nil {
			return nil, err
		}
	}
	return chartVals, nil
}

func install(config InitConfiguration) error {
	err := createNamespace(config.Namespace)
	if err != nil {
		return err
	}

	helmConf, err := helmConfig(config.Namespace)
	if err != nil {
		return err
	}

	appChart, err := appChart(config.Version, helmConf)
	if err != nil {
		return err
	}

	version, err := getVersion(config.Version)
	if err != nil {
		return err
	}

	err = applyCRDs(fmt.Sprintf("v%s", version))
	if err != nil {
		return err
	}

	installClient := helm.NewInstall(helmConf)
	installClient.ReleaseName = appReleaseName
	installClient.Namespace = config.Namespace
	installClient.Wait = config.Wait
	installClient.Timeout = time.Duration(config.Timeout) * time.Second

	values, err := chartValues(config)
	if err != nil {
		return err
	}

	if _, err = installClient.Run(appChart, values); err != nil {
		return err
	}
	return nil
}

func debugLogf(format string, v ...interface{}) {
}
