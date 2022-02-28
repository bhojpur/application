package runtime

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
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/bhojpur/service/pkg/utils/logger"

	"github.com/bhojpur/application/pkg/acl"
	global_config "github.com/bhojpur/application/pkg/config"
	env "github.com/bhojpur/application/pkg/config/env"
	"github.com/bhojpur/application/pkg/cors"
	"github.com/bhojpur/application/pkg/grpc"
	"github.com/bhojpur/application/pkg/metrics"
	"github.com/bhojpur/application/pkg/operator/client"
	"github.com/bhojpur/application/pkg/runtime/security"
	"github.com/bhojpur/application/pkg/utils"
	"github.com/bhojpur/application/pkg/version"
)

// FromFlags parses command flags and returns AppRuntime instance.
func FromFlags() (*AppRuntime, error) {
	mode := flag.String("mode", string(utils.StandaloneMode), "Runtime mode for Bhojpur Application runtime")
	appHTTPPort := flag.String("app-http-port", fmt.Sprintf("%v", DefaultAppHTTPPort), "HTTP port for Bhojpur Application API to listen on")
	appAPIListenAddresses := flag.String("app-listen-addresses", DefaultAPIListenAddress, "One or more addresses for the Bhojpur Application API to listen on, CSV limited")
	appPublicPort := flag.String("app-public-port", "", "Public port for Bhojpur Application Health and Metadata to listen on")
	appAPIGRPCPort := flag.String("app-grpc-port", fmt.Sprintf("%v", DefaultAppAPIGRPCPort), "gRPC port for the Bhojpur Application API to listen on")
	appInternalGRPCPort := flag.String("app-internal-grpc-port", "", "gRPC port for the Bhojpur Application Internal API to listen on")
	appPort := flag.String("app-port", "", "The port the application is listening on")
	profilePort := flag.String("profile-port", fmt.Sprintf("%v", DefaultProfilePort), "The port for the profile server")
	appProtocol := flag.String("app-protocol", string(HTTPProtocol), "Protocol for the application: gRPC or HTTP")
	componentsPath := flag.String("components-path", "", "Path for components directory. If empty, components will not be loaded. Self-hosted mode only")
	config := flag.String("config", "", "Path to config file, or name of a configuration object")
	appID := flag.String("app-id", "", "A unique ID for Bhojpur Application. Used for Service Discovery and state")
	controlPlaneAddress := flag.String("control-plane-address", "", "Address for a Bhojpur Application control plane")
	sentryAddress := flag.String("sentry-address", "", "Address for the Bhojpur Application Sentry CA service")
	placementServiceHostAddr := flag.String("placement-host-address", "", "Addresses for Bhojpur Application Actor Placement servers")
	allowedOrigins := flag.String("allowed-origins", cors.DefaultAllowedOrigins, "Allowed HTTP origins")
	enableProfiling := flag.Bool("enable-profiling", false, "Enable profiling")
	runtimeVersion := flag.Bool("version", false, "Prints the runtime version")
	buildInfo := flag.Bool("build-info", false, "Prints the build info")
	waitCommand := flag.Bool("wait", false, "wait for Bhojpur Application outbound ready")
	appMaxConcurrency := flag.Int("app-max-concurrency", -1, "Controls the concurrency level when forwarding requests to user code")
	enableMTLS := flag.Bool("enable-mtls", false, "Enables automatic mTLS for Bhojpur Application runtime-to-runtime communication channels")
	appSSL := flag.Bool("app-ssl", false, "Sets the URI scheme of the app to https and attempts an SSL connection")
	appHTTPMaxRequestSize := flag.Int("app-http-max-request-size", -1, "Increasing max size of request body in MB to handle uploading of big files. By default 4 MB.")
	unixDomainSocket := flag.String("unix-domain-socket", "", "Path to a unix domain socket dir mount. If specified, Bhojpur Application API servers will use Unix Domain Sockets")
	appHTTPReadBufferSize := flag.Int("app-http-read-buffer-size", -1, "Increasing max size of read buffer in KB to handle sending multi-KB headers. By default 4 KB.")
	appHTTPStreamRequestBody := flag.Bool("app-http-stream-request-body", false, "Enables request body streaming on http server")
	appGracefulShutdownSeconds := flag.Int("app-graceful-shutdown-seconds", -1, "Graceful shutdown time in seconds.")

	loggerOptions := logger.DefaultOptions()
	loggerOptions.AttachCmdFlags(flag.StringVar, flag.BoolVar)

	metricsExporter := metrics.NewExporter(metrics.DefaultMetricNamespace)

	metricsExporter.Options().AttachCmdFlags(flag.StringVar, flag.BoolVar)

	flag.Parse()

	if *runtimeVersion {
		fmt.Println(version.Version())
		os.Exit(0)
	}

	if *buildInfo {
		fmt.Printf("Version: %s\nGit Commit: %s\nGit Version: %s\n", version.Version(), version.Commit(), version.GitVersion())
		os.Exit(0)
	}

	if *waitCommand {
		waitUntilAppOutboundReady(*appHTTPPort)
		os.Exit(0)
	}

	if *appID == "" {
		return nil, errors.New("app-id parameter cannot be empty")
	}

	// Apply options to all loggers
	loggerOptions.SetAppID(*appID)
	if err := logger.ApplyOptionsToLoggers(&loggerOptions); err != nil {
		return nil, err
	}

	log.Infof("starting Bhojpur Application Runtime Engine -- version %s -- commit %s", version.Version(), version.Commit())
	log.Infof("log level set to: %s", loggerOptions.OutputLevel)

	// Initialize Bhojpur Application runtime metrics exporter
	if err := metricsExporter.Init(); err != nil {
		log.Fatal(err)
	}

	appHTTP, err := strconv.Atoi(*appHTTPPort)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing app-http-port flag")
	}

	appAPIGRPC, err := strconv.Atoi(*appAPIGRPCPort)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing app-grpc-port flag")
	}

	profPort, err := strconv.Atoi(*profilePort)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing profile-port flag")
	}

	var appInternalGRPC int
	if *appInternalGRPCPort != "" {
		appInternalGRPC, err = strconv.Atoi(*appInternalGRPCPort)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing app-internal-grpc-port")
		}
	} else {
		appInternalGRPC, err = grpc.GetFreePort()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get free port for internal grpc server")
		}
	}

	var publicPort *int
	if *appPublicPort != "" {
		port, cerr := strconv.Atoi(*appPublicPort)
		if cerr != nil {
			return nil, errors.Wrap(cerr, "error parsing app-public-port")
		}
		publicPort = &port
	}

	var applicationPort int
	if *appPort != "" {
		applicationPort, err = strconv.Atoi(*appPort)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing app-port")
		}
	}

	var maxRequestBodySize int
	if *appHTTPMaxRequestSize != -1 {
		maxRequestBodySize = *appHTTPMaxRequestSize
	} else {
		maxRequestBodySize = DefaultMaxRequestBodySize
	}

	var readBufferSize int
	if *appHTTPReadBufferSize != -1 {
		readBufferSize = *appHTTPReadBufferSize
	} else {
		readBufferSize = DefaultReadBufferSize
	}

	var gracefulShutdownDuration time.Duration
	if *appGracefulShutdownSeconds == -1 {
		gracefulShutdownDuration = defaultGracefulShutdownDuration
	} else {
		gracefulShutdownDuration = time.Duration(*appGracefulShutdownSeconds) * time.Second
	}

	placementAddresses := []string{}
	if *placementServiceHostAddr != "" {
		placementAddresses = parsePlacementAddr(*placementServiceHostAddr)
	}

	var concurrency int
	if *appMaxConcurrency != -1 {
		concurrency = *appMaxConcurrency
	}

	appPrtcl := string(HTTPProtocol)
	if *appProtocol != string(HTTPProtocol) {
		appPrtcl = *appProtocol
	}

	appAPIListenAddressList := strings.Split(*appAPIListenAddresses, ",")
	if len(appAPIListenAddressList) == 0 {
		appAPIListenAddressList = []string{DefaultAPIListenAddress}
	}
	runtimeConfig := NewRuntimeConfig(*appID, placementAddresses, *controlPlaneAddress, *allowedOrigins, *config, *componentsPath,
		appPrtcl, *mode, appHTTP, appInternalGRPC, appAPIGRPC, appAPIListenAddressList, publicPort, applicationPort, profPort, *enableProfiling, concurrency, *enableMTLS, *sentryAddress, *appSSL, maxRequestBodySize, *unixDomainSocket, readBufferSize, *appHTTPStreamRequestBody, gracefulShutdownDuration)

	// set environment variables
	// TODO - consider adding host address to runtime config and/or caching result in utils package
	host, err := utils.GetHostAddress()
	if err != nil {
		log.Warnf("failed to get host address, env variable %s will not be set", env.HostAddress)
	}

	variables := map[string]string{
		env.AppID:          *appID,
		env.AppPort:        *appPort,
		env.HostAddress:    host,
		env.SvcPort:        strconv.Itoa(appInternalGRPC),
		env.AppGRPCPort:    *appAPIGRPCPort,
		env.AppHTTPPort:    *appHTTPPort,
		env.AppMetricsPort: metricsExporter.Options().Port, // TODO - consider adding to runtime config
		env.AppProfilePort: *profilePort,
	}

	if err = setEnvVariables(variables); err != nil {
		return nil, err
	}

	var globalConfig *global_config.Configuration
	var configErr error

	if *enableMTLS || *mode == string(utils.KubernetesMode) {
		runtimeConfig.CertChain, err = security.GetCertChain()
		if err != nil {
			return nil, err
		}
	}

	var accessControlList *global_config.AccessControlList
	var namespace string
	var podName string

	if *config != "" {
		switch utils.AppMode(*mode) {
		case utils.KubernetesMode:
			client, conn, clientErr := client.GetOperatorClient(*controlPlaneAddress, security.TLSServerName, runtimeConfig.CertChain)
			if clientErr != nil {
				return nil, clientErr
			}
			defer conn.Close()
			namespace = os.Getenv("NAMESPACE")
			podName = os.Getenv("POD_NAME")
			globalConfig, configErr = global_config.LoadKubernetesConfiguration(*config, namespace, podName, client)
		case utils.StandaloneMode:
			globalConfig, _, configErr = global_config.LoadStandaloneConfiguration(*config)
		}
	}

	if configErr != nil {
		log.Fatalf("error loading configuration: %s", configErr)
	}
	if globalConfig == nil {
		log.Info("loading default configuration")
		globalConfig = global_config.LoadDefaultConfiguration()
	}

	accessControlList, err = acl.ParseAccessControlSpec(globalConfig.Spec.AccessControlSpec, string(runtimeConfig.ApplicationProtocol))
	if err != nil {
		log.Fatalf(err.Error())
	}
	return NewAppRuntime(runtimeConfig, globalConfig, accessControlList), nil
}

func setEnvVariables(variables map[string]string) error {
	for key, value := range variables {
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func parsePlacementAddr(val string) []string {
	parsed := []string{}
	p := strings.Split(val, ",")
	for _, addr := range p {
		parsed = append(parsed, strings.TrimSpace(addr))
	}
	return parsed
}
