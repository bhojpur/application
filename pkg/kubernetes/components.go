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
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
	"github.com/bhojpur/application/pkg/utils"
)

// ComponentsOutput represent a Bhojpur Application component.
type ComponentsOutput struct {
	Name    string `csv:"Name"`
	Type    string `csv:"Type"`
	Version string `csv:"VERSION"`
	Scopes  string `csv:"SCOPES"`
	Created string `csv:"CREATED"`
	Age     string `csv:"AGE"`
}

// PrintComponents prints all Bhojpur Application components.
func PrintComponents(name, outputFormat string) error {
	return writeComponents(os.Stdout, func() (*v1alpha1.ComponentList, error) {
		client, err := AppClient()
		if err != nil {
			return nil, err
		}

		list, err := client.ComponentsV1alpha1().Components(meta_v1.NamespaceAll).List(meta_v1.ListOptions{})
		// This means that the Bhojpur Application Components CRD is not installed and
		// therefore no component items exist.
		if apierrors.IsNotFound(err) {
			list = &v1alpha1.ComponentList{
				Items: []v1alpha1.Component{},
			}
		} else if err != nil {
			return nil, err
		}

		return list, nil
	}, name, outputFormat)
}

func writeComponents(writer io.Writer, getConfigFunc func() (*v1alpha1.ComponentList, error), name, outputFormat string) error {
	confs, err := getConfigFunc()
	if err != nil {
		return err
	}

	filtered := []v1alpha1.Component{}
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
		return printComponentList(writer, filtered)
	}

	return utils.PrintDetail(writer, outputFormat, filteredSpecs)
}

func printComponentList(writer io.Writer, list []v1alpha1.Component) error {
	co := []ComponentsOutput{}
	for _, c := range list {
		co = append(co, ComponentsOutput{
			Name:    c.GetName(),
			Type:    c.Spec.Type,
			Created: c.CreationTimestamp.Format("2006-01-02 15:04.05"),
			Age:     utils.GetAge(c.CreationTimestamp.Time),
			Version: c.Spec.Version,
			Scopes:  strings.Join(c.Scopes, ","),
		})
	}

	return utils.MarshalAndWriteTable(writer, co)
}
