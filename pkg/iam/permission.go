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
	"time"
)

type Permission struct {
	Action       string   `json:"action"`
	Actions      []string `json:"actions"`
	CreatedTime  string   `orm:"varchar(100)" json:"createdTime"`
	DisplayName  string   `orm:"varchar(100)" json:"displayName"`
	Effect       string   `orm:"varchar(100)" json:"effect"`
	IsEnabled    bool     `json:"isEnabled"`
	Name         string   `orm:"varchar(100)" json:"name"`
	Owner        string   `orm:"varchar(100)" json:"owner"`
	ResourceType string   `orm:"varchar(100)" json:"resourceType"`
	Resources    []string `json:"resources"`
	Roles        []string `json:"roles"`
	Users        []string `json:"users"`
}

func GetPermission() ([]*Permission, error) {
	queryMap := map[string]string{
		"owner":       authConfig.OrganizationName,
		"application": authConfig.ApplicationName,
	}
	url := getUrl("get-permissions", queryMap)
	bytes, err := doGetBytes(url)
	if err != nil {
		return nil, err
	}
	var permission []*Permission
	err = json.Unmarshal(bytes, &permission)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

func AddPermission(q Permission) (*Response, error) {
	data := Permission{
		Action:       "Read",
		Actions:      q.Actions,
		CreatedTime:  time.Now().UTC().String(),
		DisplayName:  q.Name,
		Effect:       "Allow",
		IsEnabled:    true,
		Name:         q.Name,
		Owner:        authConfig.OrganizationName,
		ResourceType: "Api",
		Resources:    q.Resources,
		Roles:        []string{},
		Users:        []string{},
	}

	postBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := doPost("add-permission", nil, postBytes, false)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
