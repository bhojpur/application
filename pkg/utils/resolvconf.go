package utils

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
	"bytes"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	// DefaultKubeClusterDomain is the default value of KubeClusterDomain.
	DefaultKubeClusterDomain = "cluster.local"
	defaultResolvPath        = "/etc/resolv.conf"
	commentMarker            = "#"
)

var searchRegexp = regexp.MustCompile(`^\s*search\s*(([^\s]+\s*)*)$`)

// GetKubeClusterDomain search KubeClusterDomain value from /etc/resolv.conf file.
func GetKubeClusterDomain() (string, error) {
	resolvContent, err := getResolvContent(defaultResolvPath)
	if err != nil {
		return "", err
	}
	return getClusterDomain(resolvContent)
}

func getClusterDomain(resolvConf []byte) (string, error) {
	var kubeClusterDomian string
	searchDomains := getResolvSearchDomains(resolvConf)
	sort.Strings(searchDomains)
	if len(searchDomains) == 0 || searchDomains[0] == "" {
		kubeClusterDomian = DefaultKubeClusterDomain
	} else {
		kubeClusterDomian = searchDomains[0]
	}
	return kubeClusterDomian, nil
}

func getResolvContent(resolvPath string) ([]byte, error) {
	return os.ReadFile(resolvPath)
}

func getResolvSearchDomains(resolvConf []byte) []string {
	var (
		domains []string
		lines   [][]byte
	)

	scanner := bufio.NewScanner(bytes.NewReader(resolvConf))
	for scanner.Scan() {
		line := scanner.Bytes()
		commentIndex := bytes.Index(line, []byte(commentMarker))
		if commentIndex == -1 {
			lines = append(lines, line)
		} else {
			lines = append(lines, line[:commentIndex])
		}
	}

	for _, line := range lines {
		match := searchRegexp.FindSubmatch(line)
		if match == nil {
			continue
		}
		domains = strings.Fields(string(match[1]))
	}

	return domains
}
