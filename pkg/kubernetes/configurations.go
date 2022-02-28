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
	"io"
	"os"
	"strconv"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/bhojpur/application/pkg/kubernetes/configuration/v1alpha1"
	"github.com/bhojpur/application/pkg/utils"
)

type configurationsOutput struct {
	Name           string `csv:"Name"`
	TracingEnabled bool   `csv:"TRACING-ENABLED"`
	MetricsEnabled bool   `csv:"METRICS-ENABLED"`
	Age            string `csv:"AGE"`
	Created        string `csv:"CREATED"`
}

type configurationDetailedOutput struct {
	Name string      `json:"name" yaml:"name"`
	Spec interface{} `json:"spec" yaml:"spec"`
}

// PrintConfigurations prints all Bhojpur Application configurations.
func PrintConfigurations(name, outputFormat string) error {
	return writeConfigurations(os.Stdout, func() (*v1alpha1.ConfigurationList, error) {
		client, err := AppClient()
		if err != nil {
			return nil, err
		}

		list, err := client.ConfigurationV1alpha1().Configurations(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
		// This means that the Bhojpur Application Configurations CRD is not installed and
		// therefore no configuration items exist.
		if apierrors.IsNotFound(err) {
			list = &v1alpha1.ConfigurationList{
				Items: []v1alpha1.Configuration{},
			}
		} else if err != nil {
			return nil, err
		}

		return list, err
	}, name, outputFormat)
}

func writeConfigurations(writer io.Writer, getConfigFunc func() (*v1alpha1.ConfigurationList, error), name, outputFormat string) error {
	confs, err := getConfigFunc()
	if err != nil {
		return err
	}

	filtered := []v1alpha1.Configuration{}
	filteredSpecs := []configurationDetailedOutput{}
	for _, c := range confs.Items {
		confName := c.GetName()
		if confName == "appsystem" {
			continue
		}

		if name == "" || strings.EqualFold(confName, name) {
			filtered = append(filtered, c)
			filteredSpecs = append(filteredSpecs, configurationDetailedOutput{
				Name: confName,
				Spec: c.Spec,
			})
		}
	}

	if outputFormat == "" || outputFormat == "list" {
		return printConfigurationList(writer, filtered)
	}

	return utils.PrintDetail(writer, outputFormat, filteredSpecs)
}

func printConfigurationList(writer io.Writer, list []v1alpha1.Configuration) error {
	co := []configurationsOutput{}
	for _, c := range list {
		co = append(co, configurationsOutput{
			TracingEnabled: tracingEnabled(c.Spec.TracingSpec),
			Name:           c.GetName(),
			MetricsEnabled: c.Spec.MetricSpec.Enabled,
			Created:        c.CreationTimestamp.Format("2006-01-02 15:04.05"),
			Age:            utils.GetAge(c.CreationTimestamp.Time),
		})
	}

	return utils.MarshalAndWriteTable(writer, co)
}

func tracingEnabled(spec v1alpha1.TracingSpec) bool {
	sr, err := strconv.ParseFloat(spec.SamplingRate, 32)
	if err != nil {
		return false
	}
	return sr > 0
}
