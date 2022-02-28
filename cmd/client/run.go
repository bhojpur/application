package cmd

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
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bhojpur/application/pkg/standalone"
	"github.com/bhojpur/application/pkg/utils"
)

var (
	appPort            int
	profilePort        int
	appID              string
	configFile         string
	port               int
	grpcPort           int
	maxConcurrency     int
	enableProfiling    bool
	logLevel           string
	protocol           string
	componentsPath     string
	appSSL             bool
	metricsPort        int
	maxRequestBodySize int
	unixDomainSocket   string
)

const (
	runtimeWaitTimeoutInSeconds = 60
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Bhojpur Application runtime and (optionally) your custom application side by side. Supported platforms: Self-hosted",
	Example: `
# Run a .NET application
appctl run --app-id myapp --app-port 5000 -- dotnet run

# Run a Java application
appctl run --app-id myapp -- java -jar myapp.jar

# Run a Node.js application that listens to port 3000
appctl run --app-id myapp --app-port 3000 -- node myapp.js

# Run a Python application
appctl run --app-id myapp -- python myapp.py

# Run sidecar only
appctl run --app-id myapp

# Run a gRPC application written in Go (listening on port 3000)
appctl run --app-id myapp --app-port 3000 --app-protocol grpc -- go run main.go
  `,
	Args: cobra.MinimumNArgs(0),
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("placement-host-address", cmd.Flags().Lookup("placement-host-address"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(utils.WhiteBold("WARNING: no application command found."))
		}

		if unixDomainSocket != "" {
			// TODO: add Windows support
			if runtime.GOOS == "windows" {
				utils.FailureStatusEvent(os.Stderr, "The unix-domain-socket option is not supported on Windows")
				os.Exit(1)
			} else {
				// use unix domain socket means no port any more
				utils.WarningStatusEvent(os.Stdout, "Unix domain sockets are currently a preview feature")
				port = 0
				grpcPort = 0
			}
		}

		output, err := standalone.Run(&standalone.RunConfig{
			AppID:              appID,
			AppPort:            appPort,
			HTTPPort:           port,
			GRPCPort:           grpcPort,
			ConfigFile:         configFile,
			Arguments:          args,
			EnableProfiling:    enableProfiling,
			ProfilePort:        profilePort,
			LogLevel:           logLevel,
			MaxConcurrency:     maxConcurrency,
			Protocol:           protocol,
			PlacementHostAddr:  viper.GetString("placement-host-address"),
			ComponentsPath:     componentsPath,
			AppSSL:             appSSL,
			MetricsPort:        metricsPort,
			MaxRequestBodySize: maxRequestBodySize,
			UnixDomainSocket:   unixDomainSocket,
		})
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, err.Error())
			return
		}

		sigCh := make(chan os.Signal, 1)
		setupShutdownNotify(sigCh)

		svcRunning := make(chan bool, 1)
		appRunning := make(chan bool, 1)

		go func() {
			var startInfo string
			if unixDomainSocket != "" {
				startInfo = fmt.Sprintf(
					"Starting the Bhojpur Application runtime with id %s. HTTP Socket: %v. gRPC Socket: %v.",
					output.AppID,
					utils.GetSocket(unixDomainSocket, output.AppID, "http"),
					utils.GetSocket(unixDomainSocket, output.AppID, "grpc"))
			} else {
				startInfo = fmt.Sprintf(
					"Starting Bhojput Application runtime with id %s. HTTP Port: %v. gRPC Port: %v",
					output.AppID,
					output.AppHTTPPort,
					output.AppGRPCPort)
			}
			utils.InfoStatusEvent(os.Stdout, startInfo)

			output.AppCMD.Stdout = os.Stdout
			output.AppCMD.Stderr = os.Stderr

			err = output.AppCMD.Start()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, err.Error())
				os.Exit(1)
			}

			go func() {
				svcErr := output.AppCMD.Wait()

				if svcErr != nil {
					utils.FailureStatusEvent(os.Stderr, "The Bhojpur Application runtime process exited with error code: %s", svcErr.Error())
				} else {
					utils.SuccessStatusEvent(os.Stdout, "Exited the Bhojpur Application runtime successfully")
				}
				sigCh <- os.Interrupt
			}()

			if appPort <= 0 {
				// If your application does not listen on port, we can check for Bhojpur Application runtime's
				// sidecar health before starting the application. Otherwise, it creates a deadlock.
				sidecarUp := true

				if unixDomainSocket != "" {
					httpSocket := utils.GetSocket(unixDomainSocket, output.AppID, "http")
					utils.InfoStatusEvent(os.Stdout, "Checking if Bhojpur Application runtime sidecar is listening on HTTP socket %v", httpSocket)
					err = utils.IsAppListeningOnSocket(httpSocket, time.Duration(runtimeWaitTimeoutInSeconds)*time.Second)
					if err != nil {
						sidecarUp = false
						utils.WarningStatusEvent(os.Stdout, "Bhojpur Application runtime sidecar is not listening on HTTP socket: %s", err.Error())
					}

					grpcSocket := utils.GetSocket(unixDomainSocket, output.AppID, "grpc")
					utils.InfoStatusEvent(os.Stdout, "Checking if Bhojpur Application runtime sidecar is listening on gRPC socket %v", grpcSocket)
					err = utils.IsAppListeningOnSocket(grpcSocket, time.Duration(runtimeWaitTimeoutInSeconds)*time.Second)
					if err != nil {
						sidecarUp = false
						utils.WarningStatusEvent(os.Stdout, "Bhojpur Application runtime sidecar is not listening on gRPC socket: %s", err.Error())
					}

				} else {
					utils.InfoStatusEvent(os.Stdout, "Checking if Bhojpur Application runtime sidecar is listening on HTTP port %v", output.AppHTTPPort)
					err = utils.IsAppListeningOnPort(output.AppHTTPPort, time.Duration(runtimeWaitTimeoutInSeconds)*time.Second)
					if err != nil {
						sidecarUp = false
						utils.WarningStatusEvent(os.Stdout, "Bhojpur Application runtime sidecar is not listening on HTTP port: %s", err.Error())
					}

					utils.InfoStatusEvent(os.Stdout, "Checking if Bhojpur Application runtime sidecar is listening on gRPC port %v", output.AppGRPCPort)
					err = utils.IsAppListeningOnPort(output.AppGRPCPort, time.Duration(runtimeWaitTimeoutInSeconds)*time.Second)
					if err != nil {
						sidecarUp = false
						utils.WarningStatusEvent(os.Stdout, "Bhojpur Application runtime sidecar is not listening on rRPC port: %s", err.Error())
					}
				}

				if sidecarUp {
					utils.InfoStatusEvent(os.Stdout, "Bhojpur Application runtime sidecar is up and running.")
				} else {
					utils.WarningStatusEvent(os.Stdout, "Bhojpur Application runtime sidecar might not be responding.")
				}
			}

			svcRunning <- true
		}()

		<-svcRunning

		go func() {
			if output.AppCMD == nil {
				appRunning <- true
				return
			}

			stdErrPipe, pipeErr := output.AppCMD.StderrPipe()
			if pipeErr != nil {
				utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Error creating stderr for App: %s", err.Error()))
				appRunning <- false
				return
			}

			stdOutPipe, pipeErr := output.AppCMD.StdoutPipe()
			if pipeErr != nil {
				utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Error creating stdout for App: %s", err.Error()))
				appRunning <- false
				return
			}

			errScanner := bufio.NewScanner(stdErrPipe)
			outScanner := bufio.NewScanner(stdOutPipe)
			go func() {
				for errScanner.Scan() {
					fmt.Println(utils.Blue(fmt.Sprintf("== APP == %s", errScanner.Text())))
				}
			}()

			go func() {
				for outScanner.Scan() {
					fmt.Println(utils.Blue(fmt.Sprintf("== APP == %s", outScanner.Text())))
				}
			}()

			err = output.AppCMD.Start()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, err.Error())
				appRunning <- false
				return
			}

			go func() {
				appErr := output.AppCMD.Wait()

				if appErr != nil {
					utils.FailureStatusEvent(os.Stderr, "The Bhojpur Application process exited with error code: %s", appErr.Error())
				} else {
					utils.SuccessStatusEvent(os.Stdout, "Exited the Bhojpur Application successfully")
				}
				sigCh <- os.Interrupt
			}()

			appRunning <- true
		}()

		appRunStatus := <-appRunning
		if !appRunStatus {
			// Start of Bhojpur Application failed, try to stop Bhojpur Application runtime and exit.
			err = output.AppCMD.Process.Kill()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Start of application failed, try to stop Bhojpur Application runtime Error: %s", err))
			} else {
				utils.SuccessStatusEvent(os.Stdout, "Start of application failed, try to stop Bhojpur Application runtime successfully")
			}
			os.Exit(1)
		}

		// Metadata API is only available if Bhojpur Application has started listening to port, so wait
		// for application to start before calling metadata API.
		err = utils.Put(output.AppHTTPPort, "cliPID", strconv.Itoa(os.Getpid()), appID, unixDomainSocket)
		if err != nil {
			utils.WarningStatusEvent(os.Stdout, "Could not update sidecar metadata for cliPID: %s", err.Error())
		}

		if output.AppCMD != nil {
			appCommand := strings.Join(args, " ")
			utils.InfoStatusEvent(os.Stdout, fmt.Sprintf("Updating metadata for application command: %s", appCommand))
			err = utils.Put(output.AppHTTPPort, "appCommand", appCommand, appID, unixDomainSocket)
			if err != nil {
				utils.WarningStatusEvent(os.Stdout, "Could not update sidecar metadata for appCommand: %s", err.Error())
			} else {
				utils.SuccessStatusEvent(os.Stdout, "You're up and running! Both the Bhojpur Application runtime and your application logs will appear here.\n")
			}
		} else {
			utils.SuccessStatusEvent(os.Stdout, "You're up and running! Bhojpur Application runtime logs will appear here.\n")
		}

		<-sigCh
		utils.InfoStatusEvent(os.Stdout, "\nterminated signal received: shutting down")

		if output.AppCMD.ProcessState == nil || !output.AppCMD.ProcessState.Exited() {
			err = output.AppCMD.Process.Kill()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Error exiting the Bhojpur Application runtime: %s", err))
			} else {
				utils.SuccessStatusEvent(os.Stdout, "Exited the Bhojpur Application runtime successfully")
			}
		}

		if output.AppCMD != nil && (output.AppCMD.ProcessState == nil || !output.AppCMD.ProcessState.Exited()) {
			err = output.AppCMD.Process.Kill()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Error exiting the Bhojpur Application: %s", err))
			} else {
				utils.SuccessStatusEvent(os.Stdout, "Exited the Bhojpur Application successfully")
			}
		}

		if unixDomainSocket != "" {
			for _, s := range []string{"http", "grpc"} {
				os.Remove(utils.GetSocket(unixDomainSocket, output.AppID, s))
			}
		}
	},
}

func init() {
	RunCmd.Flags().IntVarP(&appPort, "app-port", "p", -1, "The port your Bhojpur Application is listening on")
	RunCmd.Flags().StringVarP(&appID, "app-id", "a", "", "The id for your Bhojpur Application, used for service discovery")
	RunCmd.Flags().StringVarP(&configFile, "config", "c", standalone.DefaultConfigFilePath(), "Bhojpur Application runtime configuration file")
	RunCmd.Flags().IntVarP(&port, "app-http-port", "H", -1, "The HTTP port for Bhojpur Application runtime to listen on")
	RunCmd.Flags().IntVarP(&grpcPort, "app-grpc-port", "G", -1, "The gRPC port for Bhojpur Application runtime to listen on")
	RunCmd.Flags().BoolVar(&enableProfiling, "enable-profiling", false, "Enable pprof profiling via an HTTP endpoint")
	RunCmd.Flags().IntVarP(&profilePort, "profile-port", "", -1, "The port for the profile server to listen on")
	RunCmd.Flags().StringVarP(&logLevel, "log-level", "", "info", "The log verbosity. Valid values are: debug, info, warn, error, fatal, or panic")
	RunCmd.Flags().IntVarP(&maxConcurrency, "app-max-concurrency", "", -1, "The concurrency level of the application, otherwise is unlimited")
	RunCmd.Flags().StringVarP(&protocol, "app-protocol", "P", "http", "The protocol (gRPC or HTTP) Bhojpur Application runtime uses to talk to your application")
	RunCmd.Flags().StringVarP(&componentsPath, "components-path", "d", standalone.DefaultComponentsDirPath(), "The path for components directory")
	RunCmd.Flags().String("placement-host-address", "localhost", "The address of the placement service. Format is either <hostname> for default port or <hostname>:<port> for custom port")
	RunCmd.Flags().BoolVar(&appSSL, "app-ssl", false, "Enable https when Bhojpur Application runtime invokes the application")
	RunCmd.Flags().IntVarP(&metricsPort, "metrics-port", "M", -1, "The port of metrics on Bhojpur Application runtime")
	RunCmd.Flags().BoolP("help", "h", false, "Print this help message")
	RunCmd.Flags().IntVarP(&maxRequestBodySize, "app-http-max-request-size", "", -1, "Max size of request body in MB")
	RunCmd.Flags().StringVarP(&unixDomainSocket, "unix-domain-socket", "u", "", "Path to a unix domain socket dir. If specified, Bhojpur Application API servers will use Unix Domain Sockets")

	rootCmd.AddCommand(RunCmd)
}
