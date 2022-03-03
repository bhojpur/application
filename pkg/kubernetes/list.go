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
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bhojpur/application/pkg/utils"
)

// ListOutput represents the application ID, application port and creation time.
type ListOutput struct {
	AppID   string `csv:"APP ID"   json:"appId"   yaml:"appId"`
	AppPort string `csv:"APP PORT" json:"appPort" yaml:"appPort"`
	Age     string `csv:"AGE"      json:"age"     yaml:"age"`
	Created string `csv:"CREATED"  json:"created" yaml:"created"`
}

// List outputs all the applications.
func List() ([]ListOutput, error) {
	client, err := Client()
	if err != nil {
		return nil, err
	}

	podList, err := ListPods(client, meta_v1.NamespaceAll, nil)
	if err != nil {
		return nil, err
	}

	l := []ListOutput{}
	for _, p := range podList.Items {
		for _, c := range p.Spec.Containers {
			if c.Name == "appsvr" {
				lo := ListOutput{}
				for i, a := range c.Args {
					if a == "--app-port" {
						port := c.Args[i+1]
						lo.AppPort = port
					} else if a == "--app-id" {
						id := c.Args[i+1]
						lo.AppID = id
					}
				}
				lo.Created = p.CreationTimestamp.Format("2006-01-02 15:04.05")
				lo.Age = utils.GetAge(p.CreationTimestamp.Time)
				l = append(l, lo)
			}
		}
	}

	return l, nil
}
