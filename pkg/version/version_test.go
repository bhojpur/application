package version

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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersionsGithub(t *testing.T) {
	// Ensure a clean environment

	tests := []struct {
		Name         string
		Path         string
		ResponseBody string
		ExpectedErr  string
		ExpectedVer  string
	}{
		{
			"RC releases are skipped",
			"/no_rc",
			`[
  {
    "url": "https://api.github.com/repos/bhojpur/application/releases/44766923",
    "html_url": "https://github.com/bhojpur/application/releases/tag/v1.2.3-rc.1",
    "id": 44766926,
    "tag_name": "v1.2.3-rc.1",
    "target_commitish": "master",
    "name": "Bhojpur Application Runtime Engine v1.2.3-rc.1",
    "draft": false,
    "prerelease": false
  },
  {
    "url": "https://api.github.com/repos/bhojpur/application/releases/44766923",
    "html_url": "https://github.com/bhojpur/application/releases/tag/v1.2.2",
    "id": 44766923,
    "tag_name": "v1.2.2",
    "target_commitish": "master",
    "name": "Bhojpur Application Runtime Engine v1.2.2",
    "draft": false,
    "prerelease": false
  }
]
			`,
			"",
			"1.2.2",
		},
		{
			"Malformed JSON",
			"/malformed",
			"[",
			"unexpected end of JSON input",
			"",
		},
		{
			"Only RCs",
			"/only_rcs",
			`[
  {
    "url": "https://api.github.com/repos/bhojpur/application/releases/44766923",
    "html_url": "https://github.com/bhojpur/application/releases/tag/v1.2.3-rc.1",
    "id": 44766926,
    "tag_name": "v1.2.3-rc.1",
    "target_commitish": "master",
    "name": "Bhojpur Application Runtime Engine v1.2.3-rc.1",
    "draft": false,
    "prerelease": false
  }
]			`,
			"no releases",
			"",
		},
		{
			"Empty json",
			"/empty",
			"[]",
			"no releases",
			"",
		},
	}
	m := http.NewServeMux()
	s := http.Server{Addr: ":12345", Handler: m}

	for _, tc := range tests {
		body := tc.ResponseBody
		m.HandleFunc(tc.Path, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		})
	}

	go func() {
		s.ListenAndServe()
	}()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			version, err := GetLatestReleaseGithub(fmt.Sprintf("http://localhost:12345%s", tc.Path))
			assert.Equal(t, tc.ExpectedVer, version)
			if tc.ExpectedErr != "" {
				assert.EqualError(t, err, tc.ExpectedErr)
			}
		})
	}

	t.Run("error on 404", func(t *testing.T) {
		version, err := GetLatestReleaseGithub("http://localhost:12345/non-existant/path")
		assert.Equal(t, "", version)
		assert.EqualError(t, err, "http://localhost:12345/non-existant/path - 404 Not Found")
	})

	t.Run("error on bad addr", func(t *testing.T) {
		version, err := GetLatestReleaseGithub("http://a.super.non.existent.domain/")
		assert.Equal(t, "", version)
		assert.Error(t, err)
	})

	s.Shutdown(context.Background())
}

func TestGetVersionsHelm(t *testing.T) {
	// Ensure a clean environment

	tests := []struct {
		Name         string
		Path         string
		ResponseBody string
		ExpectedErr  string
		ExpectedVer  string
	}{
		{
			"RC releases are skipped",
			"/rcs_are_skiipped",
			`apiVersion: v1
entries:
  application:
  - apiVersion: v1
    appVersion: 1.2.3-rc.1
    created: "2021-06-17T03:13:24.179849371Z"
    description: A Helm chart for Bhojpur Application on Kubernetes
    digest: 60d8d17b58ca316cdcbdb8529cf9ba2c9e2e0834383c677cafbf99add86ee7a0
    name: application
    urls:
    - https://bhojpur.github.io/helm-charts/application-1.2.3-rc.1.tgz
    version: 1.2.3-rc.1
  - apiVersion: v1
    appVersion: 1.2.2
    created: "2021-06-17T03:13:24.179849371Z"
    description: A Helm chart for Bhojpur Application on Kubernetes
    digest: 60d8d17b58ca316cdcbdb8529cf9ba2c9e2e0834383c677cafbf99add86ee7a0
    name: application
    urls:
    - https://bhojpur.github.io/helm-charts/application-1.2.2.tgz
    version: 1.2.2      `,
			"",
			"1.2.2",
		},
		{
			"Malformed YAML",
			"/malformed",
			"[",
			"yaml: line 1: did not find expected node content",
			"",
		},
		{
			"Empty YAML",
			"/empty",
			"",
			"no releases",
			"",
		},
		{
			"Only RCs",
			"/only_rcs",
			`apiVersion: v1
entries:
  application:
  - apiVersion: v1
    appVersion: 1.2.3-rc.1
    created: "2021-06-17T03:13:24.179849371Z"
    description: A Helm chart for Bhojpur Application on Kubernetes
    digest: 60d8d17b58ca316cdcbdb8529cf9ba2c9e2e0834383c677cafbf99add86ee7a0
    name: application
    urls:
    - https://bhojpur.github.io/helm-charts/application-1.2.3-rc.1.tgz
    version: 1.2.3-rc.1 `,
			"no releases",
			"",
		},
	}
	m := http.NewServeMux()
	s := http.Server{Addr: ":12346", Handler: m}

	for _, tc := range tests {
		body := tc.ResponseBody
		m.HandleFunc(tc.Path, func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		})
	}

	go func() {
		s.ListenAndServe()
	}()

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			version, err := GetLatestReleaseHelmChart(fmt.Sprintf("http://localhost:12346%s", tc.Path))
			assert.Equal(t, tc.ExpectedVer, version)
			if tc.ExpectedErr != "" {
				assert.EqualError(t, err, tc.ExpectedErr)
			}
		})
	}

	s.Shutdown(context.Background())
}
