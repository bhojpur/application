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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/bhojpur/application/pkg/utils"
)

var (
	// Version is the semver release name of this build
	version string = "edge"
	// Commit is the commit hash this build was created from
	commit string
	// Date is the time when this build was created
	date string

	gitcommit, gitversion string
)

// Print writes the version info to stdout
func Print() {
	fmt.Printf("Version:    %s\n", version)
	fmt.Printf("Commit:     %s\n", commit)
	fmt.Printf("Build Date: %s\n", date)
}

const (
	// AppGitHubOrg is the org name of Bhojpur Application on GitHub.
	AppGitHubOrg = "bhojpur"
	// AppGitHubRepo is the repo name of Bhojpur Application runtime on GitHub.
	AppGitHubRepo = "application"
	// DashboardGitHubRepo is the repo name of Bhojpur Application dashboard on GitHub.
	DashboardGitHubRepo = "dashboard"
)

type githubRepoReleaseItem struct {
	URL     string `json:"url"`
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Draft   bool   `json:"draft"`
}

type helmChartItems struct {
	Entries struct {
		App []struct {
			Version string `yaml:"appVersion"`
		}
	}
}

func GetDashboardVersion() (string, error) {
	return GetLatestReleaseGithub(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", AppGitHubOrg, DashboardGitHubRepo))
}

func GetAppVersion() (string, error) {
	version, err := GetLatestReleaseGithub(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", AppGitHubOrg, AppGitHubRepo))
	if err != nil {
		utils.WarningStatusEvent(os.Stdout, "Failed to get runtime version: '%s'. Trying secondary source", err)

		version, err = GetLatestReleaseHelmChart("https://app.github.io/helm-charts/index.yaml")
		if err != nil {
			return "", err
		}
	}

	return version, nil
}

func GetVersionFromURL(releaseURL string, parseVersion func(body []byte) (string, error)) (string, error) {
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return "", err
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken != "" {
		req.Header.Add("Authorization", "token "+githubToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s - %s", releaseURL, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return parseVersion(body)
}

// GetLatestReleaseGithub return the latest release version of Bhojpur Application from GitHub API.
func GetLatestReleaseGithub(githubURL string) (string, error) {
	return GetVersionFromURL(githubURL, func(body []byte) (string, error) {
		var githubRepoReleases []githubRepoReleaseItem
		err := json.Unmarshal(body, &githubRepoReleases)
		if err != nil {
			return "", err
		}

		if len(githubRepoReleases) == 0 {
			return "", fmt.Errorf("no releases")
		}

		for _, release := range githubRepoReleases {
			if !strings.Contains(release.TagName, "-rc") {
				return strings.TrimPrefix(release.TagName, "v"), nil
			}
		}

		return "", fmt.Errorf("no releases")
	})
}

// GetLatestReleaseHelmChart return the latest release version of Bhojpur Application
// from helm chart static index.yaml.
func GetLatestReleaseHelmChart(helmChartURL string) (string, error) {
	return GetVersionFromURL(helmChartURL, func(body []byte) (string, error) {
		var helmChartReleases helmChartItems
		err := yaml.Unmarshal(body, &helmChartReleases)
		if err != nil {
			return "", err
		}
		if len(helmChartReleases.Entries.App) == 0 {
			return "", fmt.Errorf("no releases")
		}

		for _, release := range helmChartReleases.Entries.App {
			if !strings.Contains(release.Version, "-rc") {
				return release.Version, nil
			}
		}

		return "", fmt.Errorf("no releases")
	})
}

// Version returns the Bhojpur Application runtime version. This is either a
// semantic version number or else, in the case of unreleased code, the string "edge".
func Version() string {
	return version
}

// Commit returns the git commit SHA for the code that Bhojpur Application was built from.
func Commit() string {
	return gitcommit
}

// Date returns the build date
func Date() string {
	return date
}

// GitVersion returns the git version for the code that Bhojpur Application runtime
// was built from.
func GitVersion() string {
	return gitversion
}
