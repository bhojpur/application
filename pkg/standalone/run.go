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
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	"github.com/Pallinder/sillyname-go"
	"github.com/phayes/freeport"
	"gopkg.in/yaml.v2"

	"github.com/bhojpur/application/pkg/components"
	modes "github.com/bhojpur/application/pkg/config/modes"
)

const sentryDefaultAddress = "localhost:50001"

// RunConfig represents the application configuration parameters.
type RunConfig struct {
	AppID              string `env:"APP_ID" arg:"app-id"`
	AppPort            int    `env:"APP_PORT" arg:"app-port"`
	HTTPPort           int    `env:"APP_HTTP_PORT" arg:"app-http-port"`
	GRPCPort           int    `env:"APP_GRPC_PORT" arg:"app-grpc-port"`
	ConfigFile         string `arg:"config"`
	Protocol           string `arg:"app-protocol"`
	Arguments          []string
	EnableProfiling    bool   `arg:"enable-profiling"`
	ProfilePort        int    `arg:"profile-port"`
	LogLevel           string `arg:"log-level"`
	MaxConcurrency     int    `arg:"app-max-concurrency"`
	PlacementHostAddr  string `arg:"placement-host-address"`
	ComponentsPath     string `arg:"components-path"`
	AppSSL             bool   `arg:"app-ssl"`
	MetricsPort        int    `env:"APP_METRICS_PORT" arg:"metrics-port"`
	MaxRequestBodySize int    `arg:"app-http-max-request-size"`
	UnixDomainSocket   string `arg:"unix-domain-socket"`
}

func (meta *AppMeta) newAppID() string {
	for {
		appID := strings.ReplaceAll(sillyname.GenerateStupidName(), " ", "-")
		if !meta.idExists(appID) {
			return appID
		}
	}
}

func (config *RunConfig) validateComponentPath() error {
	_, err := os.Stat(config.ComponentsPath)
	if err != nil {
		return err
	}
	componentsLoader := components.NewStandaloneComponents(modes.StandaloneConfig{ComponentsPath: config.ComponentsPath})
	_, err = componentsLoader.LoadComponents()
	if err != nil {
		return err
	}
	return nil
}

func (config *RunConfig) validatePlacementHostAddr() error {
	placementHostAddr := config.PlacementHostAddr
	if len(placementHostAddr) == 0 {
		placementHostAddr = "localhost"
	}
	if indx := strings.Index(placementHostAddr, ":"); indx == -1 {
		if runtime.GOOS == appWindowsOS {
			placementHostAddr = fmt.Sprintf("%s:6050", placementHostAddr)
		} else {
			placementHostAddr = fmt.Sprintf("%s:50005", placementHostAddr)
		}
	}
	config.PlacementHostAddr = placementHostAddr
	return nil
}

func (config *RunConfig) validatePort(portName string, portPtr *int, meta *AppMeta) error {
	if *portPtr <= 0 {
		port, err := freeport.GetFreePort()
		if err != nil {
			return err
		}
		*portPtr = port
		return nil
	}

	if meta.portExists(*portPtr) {
		return fmt.Errorf("invalid configuration for %s. Port %v is not available", portName, *portPtr)
	}
	return nil
}

func (config *RunConfig) validate() error {
	meta, err := newAppMeta()
	if err != nil {
		return err
	}

	if config.AppID == "" {
		config.AppID = meta.newAppID()
	}

	err = config.validateComponentPath()
	if err != nil {
		return err
	}

	if config.AppPort < 0 {
		config.AppPort = 0
	}

	err = config.validatePort("HTTPPort", &config.HTTPPort, meta)
	if err != nil {
		return err
	}

	err = config.validatePort("GRPCPort", &config.GRPCPort, meta)
	if err != nil {
		return err
	}

	err = config.validatePort("MetricsPort", &config.MetricsPort, meta)
	if err != nil {
		return err
	}

	if config.EnableProfiling {
		err = config.validatePort("ProfilePort", &config.ProfilePort, meta)
		if err != nil {
			return err
		}
	}

	if config.MaxConcurrency < 1 {
		config.MaxConcurrency = -1
	}
	if config.MaxRequestBodySize < 0 {
		config.MaxRequestBodySize = -1
	}

	err = config.validatePlacementHostAddr()
	if err != nil {
		return err
	}
	return nil
}

type AppMeta struct {
	ExistingIDs   map[string]bool
	ExistingPorts map[int]bool
}

func (meta *AppMeta) idExists(id string) bool {
	_, ok := meta.ExistingIDs[id]
	return ok
}

func (meta *AppMeta) portExists(port int) bool {
	if port <= 0 {
		return false
	}
	_, ok := meta.ExistingPorts[port]
	if ok {
		return true
	}

	// try to listen on the port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return true
	}
	listener.Close()

	meta.ExistingPorts[port] = true
	return false
}

func newAppMeta() (*AppMeta, error) {
	meta := AppMeta{}
	meta.ExistingIDs = make(map[string]bool)
	meta.ExistingPorts = make(map[int]bool)
	app, err := List()
	if err != nil {
		return nil, err
	}
	for _, instance := range app {
		meta.ExistingIDs[instance.AppID] = true
		meta.ExistingPorts[instance.AppPort] = true
		meta.ExistingPorts[instance.HTTPPort] = true
		meta.ExistingPorts[instance.GRPCPort] = true
	}
	return &meta, nil
}

func (config *RunConfig) getArgs() []string {
	args := []string{}
	schema := reflect.ValueOf(*config)
	for i := 0; i < schema.NumField(); i++ {
		valueField := schema.Field(i).Interface()
		typeField := schema.Type().Field(i)
		key := typeField.Tag.Get("arg")
		if len(key) == 0 {
			continue
		}
		key = "--" + key

		switch valueField.(type) {
		case bool:
			if valueField == true {
				args = append(args, key)
			}
		default:
			value := fmt.Sprintf("%v", reflect.ValueOf(valueField))
			if len(value) != 0 {
				args = append(args, key, value)
			}
		}
	}
	if config.ConfigFile != "" {
		sentryAddress := mtlsEndpoint(config.ConfigFile)
		if sentryAddress != "" {
			// mTLS is enabled locally, set it up
			args = append(args, "--enable-mtls", "--sentry-address", sentryAddress)
		}
	}

	return args
}

func (config *RunConfig) getEnv() []string {
	env := []string{}
	schema := reflect.ValueOf(*config)
	for i := 0; i < schema.NumField(); i++ {
		valueField := schema.Field(i).Interface()
		typeField := schema.Type().Field(i)
		key := typeField.Tag.Get("env")
		if len(key) == 0 {
			continue
		}
		if value, ok := valueField.(int); ok && value <= 0 {
			// ignore unset numeric variables
			continue
		}

		value := fmt.Sprintf("%v", reflect.ValueOf(valueField))
		env = append(env, fmt.Sprintf("%s=%v", key, value))
	}
	return env
}

// RunOutput represents the run output.
type RunOutput struct {
	SvrCMD      *exec.Cmd
	AppHTTPPort int
	AppGRPCPort int
	AppID       string
	AppCMD      *exec.Cmd
}

func getSvrCommand(config *RunConfig) (*exec.Cmd, error) {
	appCMD := binaryFilePath(defaultAppBinPath(), "appsvr")
	args := config.getArgs()
	cmd := exec.Command(appCMD, args...)
	return cmd, nil
}

func mtlsEndpoint(configFile string) string {
	if configFile == "" {
		return ""
	}

	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		return ""
	}

	var config mtlsConfig
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return ""
	}

	if config.Spec.MTLS.Enabled {
		return sentryDefaultAddress
	}
	return ""
}

func getAppCommand(config *RunConfig) *exec.Cmd {
	argCount := len(config.Arguments)

	if argCount == 0 {
		return nil
	}
	command := config.Arguments[0]

	args := []string{}
	if argCount > 1 {
		args = config.Arguments[1:]
	}

	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, config.getEnv()...)

	return cmd
}

func Run(config *RunConfig) (*RunOutput, error) {
	err := config.validate()
	if err != nil {
		return nil, err
	}

	svrCMD, err := getSvrCommand(config)
	if err != nil {
		return nil, err
	}

	var appCMD *exec.Cmd = getAppCommand(config)
	return &RunOutput{
		SvrCMD:      svrCMD,
		AppCMD:      appCMD,
		AppID:       config.AppID,
		AppHTTPPort: config.HTTPPort,
		AppGRPCPort: config.GRPCPort,
	}, nil
}
