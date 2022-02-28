package standalone

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
	"bufio"
	"os/exec"
	"strings"
)

// Values for these are injected by the build.
var (
	gitcommit, gitversion string
)

// GetRuntimeVersion returns the version for the local Bhojpur Application runtime.
func GetRuntimeVersion() string {
	appBinDir := defaultAppBinPath()
	appCMD := binaryFilePath(appBinDir, "appsvr")

	out, err := exec.Command(appCMD, "--version").Output()
	if err != nil {
		return "n/a\n"
	}
	return string(out)
}

// GetDashboardVersion returns the version for the local Bhojpur Dashboard.
func GetDashboardVersion() string {
	appBinDir := defaultAppBinPath()
	dashboardCMD := binaryFilePath(appBinDir, "dashboard")

	out, err := exec.Command(dashboardCMD, "--version").Output()
	if err != nil {
		return "n/a\n"
	}
	return string(out)
}

// GetBuildInfo returns build info for the CLI and the local Bhojpur Application runtime.
func GetBuildInfo(version string) string {
	appBinDir := defaultAppBinPath()
	appCMD := binaryFilePath(appBinDir, "appsvr")

	strs := []string{
		"CLI:",
		"\tVersion: " + version,
		"\tGit Commit: " + gitcommit,
		"\tGit Version: " + gitversion,
		"Runtime:",
	}

	out, err := exec.Command(appCMD, "--build-info").Output()
	if err != nil {
		// try '--version' for older runtime version
		out, err = exec.Command(appCMD, "--version").Output()
	}
	if err != nil {
		strs = append(strs, "\tN/A")
	} else {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			strs = append(strs, "\t"+scanner.Text())
		}
	}
	return strings.Join(strs, "\n")
}
