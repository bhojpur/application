package components

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
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config "github.com/bhojpur/application/pkg/config/modes"
	components_v1alpha1 "github.com/bhojpur/application/pkg/kubernetes/components/v1alpha1"
)

const (
	yamlSeparator = "\n---"
	componentKind = "Component"
)

// StandaloneComponents loads components in a standalone mode environment.
type StandaloneComponents struct {
	config config.StandaloneConfig
}

// NewStandaloneComponents returns a new standalone loader.
func NewStandaloneComponents(configuration config.StandaloneConfig) *StandaloneComponents {
	return &StandaloneComponents{
		config: configuration,
	}
}

// LoadComponents loads Bhojpur Application components from a given directory.
func (s *StandaloneComponents) LoadComponents() ([]components_v1alpha1.Component, error) {
	files, err := os.ReadDir(s.config.ComponentsPath)
	if err != nil {
		return nil, err
	}

	list := []components_v1alpha1.Component{}

	for _, file := range files {
		if !file.IsDir() && s.isYaml(file.Name()) {
			components := s.loadComponentsFromFile(file.Name())
			if len(components) > 0 {
				list = append(list, components...)
			}
		}
	}

	return list, nil
}

func (s *StandaloneComponents) loadComponentsFromFile(filename string) []components_v1alpha1.Component {
	var errors []error

	components := []components_v1alpha1.Component{}
	path := filepath.Join(s.config.ComponentsPath, filename)

	b, err := os.ReadFile(path)
	if err != nil {
		log.Warnf("Bhojpur Application runtime load components error when reading file %s : %s", path, err)
		return components
	}
	components, errors = s.decodeYaml(b)
	for _, err := range errors {
		log.Warnf("Bhojpur Application runtime load components error when parsing components yaml resource in %s : %s", path, err)
	}
	return components
}

// isYaml checks whether the file is yaml or not.
func (s *StandaloneComponents) isYaml(fileName string) bool {
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension == ".yaml" || extension == ".yml" {
		return true
	}
	return false
}

// decodeYaml decodes the yaml document.
func (s *StandaloneComponents) decodeYaml(b []byte) ([]components_v1alpha1.Component, []error) {
	list := []components_v1alpha1.Component{}
	errors := []error{}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(s.splitYamlDoc)

	type typeInfo struct {
		metav1.TypeMeta `json:",inline"`
	}

	for {
		if !scanner.Scan() {
			err := scanner.Err()
			if err != nil {
				errors = append(errors, err)

				continue
			}

			break
		}

		scannerBytes := scanner.Bytes()
		var ti typeInfo
		if err := yaml.Unmarshal(scannerBytes, &ti); err != nil {
			errors = append(errors, err)

			continue
		}

		if ti.Kind != componentKind {
			continue
		}

		var comp components_v1alpha1.Component
		comp.Spec = components_v1alpha1.ComponentSpec{}
		if err := yaml.Unmarshal(scannerBytes, &comp); err != nil {
			errors = append(errors, err)

			continue
		}

		list = append(list, comp)
	}

	return list, errors
}

// splitYamlDoc - splits the yaml docs.
func (s *StandaloneComponents) splitYamlDoc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		i += sep
		after := data[i:]

		if len(after) == 0 {
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
