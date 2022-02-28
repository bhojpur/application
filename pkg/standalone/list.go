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
	"os"
	"strconv"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
	process "github.com/shirou/gopsutil/process"

	"github.com/bhojpur/application/pkg/utils"
)

// ListOutput represents the application ID, application port and creation time.
type ListOutput struct {
	AppID          string `csv:"APP ID"    json:"appId"          yaml:"appId"`
	HTTPPort       int    `csv:"HTTP PORT" json:"httpPort"       yaml:"httpPort"`
	GRPCPort       int    `csv:"GRPC PORT" json:"grpcPort"       yaml:"grpcPort"`
	AppPort        int    `csv:"APP PORT"  json:"appPort"        yaml:"appPort"`
	MetricsEnabled bool   `csv:"-"         json:"metricsEnabled" yaml:"metricsEnabled"` // Not displayed in table, consumed by dashboard.
	Command        string `csv:"COMMAND"   json:"command"        yaml:"command"`
	Age            string `csv:"AGE"       json:"age"            yaml:"age"`
	Created        string `csv:"CREATED"   json:"created"        yaml:"created"`
	PID            int    `csv:"PID"       json:"pid"            yaml:"pid"`
}

// runData is a placeholder for collected information linking cli and sidecar.
type runData struct {
	cliPID             int
	sidecarPID         int
	grpcPort           int
	httpPort           int
	appPort            int
	appID              string
	appCmd             string
	enableMetrics      bool
	maxRequestBodySize int
}

func (d *appProcess) List() ([]ListOutput, error) {
	return List()
}

// List outputs all the applications.
func List() ([]ListOutput, error) {
	list := []ListOutput{}

	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}

	// Links a cli PID to the corresponding sidecar Process.
	cliToSidecarMap := make(map[int]*runData)

	// Populates the map if all data is available for the sidecar.
	for _, proc := range processes {
		executable := strings.ToLower(proc.Executable())
		if (executable == "appctl") || (executable == "appctl.exe") {
			procDetails, err := process.NewProcess(int32(proc.Pid()))
			if err != nil {
				continue
			}

			cmdLine, err := procDetails.Cmdline()
			if err != nil {
				continue
			}

			cmdLineItems := strings.Fields(cmdLine)
			if len(cmdLineItems) <= 1 {
				continue
			}

			argumentsMap := make(map[string]string)
			for i := 1; i < len(cmdLineItems)-1; i += 2 {
				argumentsMap[cmdLineItems[i]] = cmdLineItems[i+1]
			}

			httpPort, err := strconv.Atoi(argumentsMap["--app-http-port"])
			if err != nil {
				continue
			}

			grpcPort, err := strconv.Atoi(argumentsMap["--app-grpc-port"])
			if err != nil {
				continue
			}

			appPort, err := strconv.Atoi(argumentsMap["--app-port"])
			if err != nil {
				appPort = 0
			}

			enableMetrics, err := strconv.ParseBool(argumentsMap["--enable-metrics"])
			if err != nil {
				// Default is true for metrics.
				enableMetrics = true
			}
			appID := argumentsMap["--app-id"]
			appCmd := ""
			cliPIDString := ""
			socket := argumentsMap["--unix-domain-socket"]
			appMetadata, err := utils.Get(httpPort, appID, socket)
			if err == nil {
				appCmd = appMetadata.Extended["appCommand"]
				cliPIDString = appMetadata.Extended["cliPID"]
			}

			// Parse functions return an error on bad input.
			cliPID, err := strconv.Atoi(cliPIDString)
			if err != nil {
				continue
			}

			maxRequestBodySize, err := strconv.Atoi(argumentsMap["--app-http-max-request-size"])
			if err != nil {
				continue
			}

			run := runData{
				cliPID:             cliPID,
				sidecarPID:         proc.Pid(),
				grpcPort:           grpcPort,
				httpPort:           httpPort,
				appPort:            appPort,
				appID:              appID,
				appCmd:             appCmd,
				enableMetrics:      enableMetrics,
				maxRequestBodySize: maxRequestBodySize,
			}

			cliToSidecarMap[cliPID] = &run
		}
	}

	myPID := os.Getpid()
	// The master list comes from cli processes, even if sidecar is not up.
	for _, proc := range processes {
		executable := strings.ToLower(proc.Executable())
		if (executable == "appctl") || (executable == "appctl.exe") {
			pID := proc.Pid()
			if pID == myPID {
				// Do not display current `appctl list` process.
				continue
			}

			procDetails, err := process.NewProcess(int32(pID))
			if err != nil {
				continue
			}

			createUnixTimeMilliseconds, err := procDetails.CreateTime()
			if err != nil {
				continue
			}

			createTime := time.Unix(createUnixTimeMilliseconds/1000, 0)

			listRow := ListOutput{
				Created: createTime.Format("2006-01-02 15:04.05"),
				Age:     utils.GetAge(createTime),
				PID:     proc.Pid(),
			}

			// Now we use sidecar into to decorate with more info (empty, if sidecar is down).
			run, ok := cliToSidecarMap[proc.Pid()]
			if ok {
				listRow.AppID = run.appID
				listRow.HTTPPort = run.httpPort
				listRow.GRPCPort = run.grpcPort
				listRow.AppPort = run.appPort
				listRow.MetricsEnabled = run.enableMetrics
				listRow.Command = utils.TruncateString(run.appCmd, 20)
			}

			// filter only dashboard instance
			if listRow.AppID != "" {
				list = append(list, listRow)
			}
		}
	}

	return list, nil
}
