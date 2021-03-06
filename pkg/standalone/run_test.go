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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertArgumentEqual(t *testing.T, key string, expectedValue string, args []string) {
	var value string
	for index, arg := range args {
		if arg == "--"+key {
			nextIndex := index + 1
			if nextIndex < len(args) {
				if !strings.HasPrefix(args[nextIndex], "--") {
					value = args[nextIndex]
				}
			}
		}
	}

	assert.Equal(t, expectedValue, value)
}

func assertArgumentNotEqual(t *testing.T, key string, expectedValue string, args []string) {
	var value string
	for index, arg := range args {
		if arg == "--"+key {
			nextIndex := index + 1
			if nextIndex < len(args) {
				if !strings.HasPrefix(args[nextIndex], "--") {
					value = args[nextIndex]
				}
			}
		}
	}

	assert.NotEqual(t, expectedValue, value)
}

func setupRun(t *testing.T) {
	componentsDir := DefaultComponentsDirPath()
	configFile := DefaultConfigFilePath()
	err := os.MkdirAll(componentsDir, 0o700)
	assert.Equal(t, nil, err, "Unable to setup components dir before running test")
	file, err := os.Create(configFile)
	file.Close()
	assert.Equal(t, nil, err, "Unable to create config file before running test")
}

func tearDownRun(t *testing.T) {
	err := os.RemoveAll(DefaultComponentsDirPath())
	assert.Equal(t, nil, err, "Unable to delete default components dir after running test")
	err = os.Remove(DefaultConfigFilePath())
	assert.Equal(t, nil, err, "Unable to delete default config file after running test")
}

func assertCommonArgs(t *testing.T, basicConfig *RunConfig, output *RunOutput) {
	assert.NotNil(t, output)

	assert.Equal(t, "MyID", output.AppID)
	assert.Equal(t, 8000, output.AppHTTPPort)
	assert.Equal(t, 50001, output.AppGRPCPort)

	assert.Contains(t, output.SvrCMD.Args[0], "appsvr")
	assertArgumentEqual(t, "app-id", "MyID", output.AppCMD.Args)
	assertArgumentEqual(t, "app-http-port", "8000", output.AppCMD.Args)
	assertArgumentEqual(t, "app-grpc-port", "50001", output.AppCMD.Args)
	assertArgumentEqual(t, "log-level", basicConfig.LogLevel, output.AppCMD.Args)
	assertArgumentEqual(t, "app-max-concurrency", "-1", output.AppCMD.Args)
	assertArgumentEqual(t, "app-protocol", "http", output.AppCMD.Args)
	assertArgumentEqual(t, "app-port", "3000", output.AppCMD.Args)
	assertArgumentEqual(t, "components-path", DefaultComponentsDirPath(), output.AppCMD.Args)
	assertArgumentEqual(t, "app-ssl", "", output.AppCMD.Args)
	assertArgumentEqual(t, "metrics-port", "9001", output.AppCMD.Args)
	assertArgumentEqual(t, "app-http-max-request-size", "-1", output.AppCMD.Args)
}

func assertAppEnv(t *testing.T, config *RunConfig, output *RunOutput) {
	envSet := make(map[string]bool)
	for _, env := range output.AppCMD.Env {
		envSet[env] = true
	}

	expectedEnvSet := getEnvSet(config)
	for _, env := range expectedEnvSet {
		_, found := envSet[env]
		if !found {
			assert.Fail(t, "Missing environment variable. Expected to have "+env)
		}
	}
}

func getEnvSet(config *RunConfig) []string {
	set := []string{
		getEnv("APP_GRPC_PORT", config.GRPCPort),
		getEnv("APP_HTTP_PORT", config.HTTPPort),
		getEnv("APP_METRICS_PORT", config.MetricsPort),
		getEnv("APP_ID", config.AppID),
	}
	if config.AppPort > 0 {
		set = append(set, getEnv("APP_PORT", config.AppPort))
	}
	if config.EnableProfiling {
		set = append(set, getEnv("APP_PROFILE_PORT", config.ProfilePort))
	}
	return set
}

func getEnv(key string, value interface{}) string {
	return fmt.Sprintf("%s=%v", key, value)
}

func TestRun(t *testing.T) {
	// Setup the components directory which is done at init time
	setupRun(t)

	// Setup the tearDown routine to run in the end
	defer tearDownRun(t)

	basicConfig := &RunConfig{
		AppID:              "MyID",
		AppPort:            3000,
		HTTPPort:           8000,
		GRPCPort:           50001,
		LogLevel:           "WARN",
		Arguments:          []string{"MyCommand", "--my-arg"},
		EnableProfiling:    false,
		ProfilePort:        9090,
		Protocol:           "http",
		ComponentsPath:     DefaultComponentsDirPath(),
		AppSSL:             true,
		MetricsPort:        9001,
		MaxRequestBodySize: -1,
	}

	t.Run("run happy http", func(t *testing.T) {
		output, err := Run(basicConfig)
		assert.Nil(t, err)

		assertCommonArgs(t, basicConfig, output)
		assert.Equal(t, "MyCommand", output.AppCMD.Args[0])
		assert.Equal(t, "--my-arg", output.AppCMD.Args[1])
		assertAppEnv(t, basicConfig, output)
	})

	t.Run("run without app command", func(t *testing.T) {
		basicConfig.Arguments = nil
		basicConfig.LogLevel = "INFO"
		basicConfig.ConfigFile = DefaultConfigFilePath()
		output, err := Run(basicConfig)
		assert.Nil(t, err)

		assertCommonArgs(t, basicConfig, output)
		assertArgumentEqual(t, "config", DefaultConfigFilePath(), output.AppCMD.Args)
		assert.Nil(t, output.AppCMD)
	})

	t.Run("run without port", func(t *testing.T) {
		basicConfig.HTTPPort = -1
		basicConfig.GRPCPort = -1
		basicConfig.MetricsPort = -1
		output, err := Run(basicConfig)

		assert.Nil(t, err)
		assert.NotNil(t, output)

		assertArgumentNotEqual(t, "http-port", "-1", output.AppCMD.Args)
		assertArgumentNotEqual(t, "grpc-port", "-1", output.AppCMD.Args)
		assertArgumentNotEqual(t, "metrics-port", "-1", output.AppCMD.Args)
	})
}
