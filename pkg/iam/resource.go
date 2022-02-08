package iam

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
	"encoding/json"
	"fmt"
)

// Resource has the same definition as Bhojpur IAM resource object
type Resource struct {
	Owner string `orm:"varchar(100) notnull pk" json:"owner"`
	Name  string `orm:"varchar(100) notnull pk" json:"name"`
}

func UploadResource(user string, tag string, parent string, fullFilePath string, fileBytes []byte) (string, string, error) {
	queryMap := map[string]string{
		"owner":        authConfig.OrganizationName,
		"user":         user,
		"application":  authConfig.ApplicationName,
		"tag":          tag,
		"parent":       parent,
		"fullFilePath": fullFilePath,
	}

	resp, err := doPost("upload-resource", queryMap, fileBytes, true)
	if err != nil {
		return "", "", err
	}

	if resp.Status != "ok" {
		return "", "", fmt.Errorf(resp.Msg)
	}

	fileUrl := resp.Data.(string)
	name := resp.Data2.(string)
	return fileUrl, name, nil
}

func UploadResourceEx(user string, tag string, parent string, fullFilePath string, fileBytes []byte, createdTime string, description string) (string, string, error) {
	queryMap := map[string]string{
		"owner":        authConfig.OrganizationName,
		"user":         user,
		"application":  authConfig.ApplicationName,
		"tag":          tag,
		"parent":       parent,
		"fullFilePath": fullFilePath,
		"createdTime":  createdTime,
		"description":  description,
	}

	resp, err := doPost("upload-resource", queryMap, fileBytes, true)
	if err != nil {
		return "", "", err
	}

	if resp.Status != "ok" {
		return "", "", fmt.Errorf(resp.Msg)
	}

	fileUrl := resp.Data.(string)
	name := resp.Data2.(string)
	return fileUrl, name, nil
}

func DeleteResource(name string) (bool, error) {
	resource := Resource{
		Owner: authConfig.OrganizationName,
		Name:  name,
	}
	postBytes, err := json.Marshal(resource)
	if err != nil {
		return false, err
	}

	resp, err := doPost("delete-resource", nil, postBytes, false)
	if err != nil {
		return false, err
	}

	return resp.Data == "Affected", nil
}
