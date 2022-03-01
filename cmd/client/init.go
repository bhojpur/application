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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bhojpur/application/pkg/kubernetes"
	"github.com/bhojpur/application/pkg/standalone"
	"github.com/bhojpur/application/pkg/utils"
)

var (
	kubernetesMode   bool
	wait             bool
	timeout          uint
	slimMode         bool
	runtimeVersion   string
	dashboardVersion string
	initNamespace    string
	enableMTLS       bool
	enableHA         bool
	values           []string
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Install the runtime on supported hosting platforms. Supported platforms: Kubernetes and Self-hosted",
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("network", cmd.Flags().Lookup("network"))
		viper.BindPFlag("image-repository", cmd.Flags().Lookup("image-repository"))
	},
	Example: `
# Initialize the Bhojpur Application runtime in self-hosted mode
appctl init

#Initialize the Bhojpur Application runtime in self-hosted mode with a provided Docker image repository. Image looked up as <repository-url>/<image>
appctl init --image-repository <repository-url>

# Initialize the Bhojpur Application runtime in Kubernetes
appctl init -k

# Initialize the Bhojpur Application runtime in Kubernetes and wait for the installation to complete (default timeout is 300s/5m)
appctl init -k --wait --timeout 600

# Initialize particular Bhojpur Application runtime in self-hosted mode
appctl init --runtime-version 0.10.0

# Initialize particular Bhojpur Application runtime in Kubernetes
appctl init -k --runtime-version 0.10.0

# Initialize the Bhojpur Application in slim self-hosted mode
appctl init -s

# See more at: https://docs.bhojpur.net/getting-started/
`,
	Run: func(cmd *cobra.Command, args []string) {
		utils.PendingStatusEvent(os.Stdout, "Making the jump to the Bhojpur.NET platform...")

		if kubernetesMode {
			utils.InfoStatusEvent(os.Stdout, "Note: To install Bhojpur Application using Helm, see here: https://docs.bhojur.net/getting-started/install-on-kubernetes/#install-with-helm-advanced\n")

			config := kubernetes.InitConfiguration{
				Namespace:  initNamespace,
				Version:    runtimeVersion,
				EnableMTLS: enableMTLS,
				EnableHA:   enableHA,
				Args:       values,
				Wait:       wait,
				Timeout:    timeout,
			}
			err := kubernetes.Init(config)
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, err.Error())
				os.Exit(1)
			}
			utils.SuccessStatusEvent(os.Stdout, fmt.Sprintf("Success! Bhojpur Application runtime has been installed to namespace %s. To verify, run `appctl status -k' in your terminal. To get started, go here: https://aka.ms/bhojpur-getting-started", config.Namespace))
		} else {
			dockerNetwork := ""
			imageRepositoryURL := ""
			if !slimMode {
				dockerNetwork = viper.GetString("network")
				imageRepositoryURL = viper.GetString("image-repository")
			}
			err := standalone.Init(runtimeVersion, dashboardVersion, dockerNetwork, slimMode, imageRepositoryURL)
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, err.Error())
				os.Exit(1)
			}
			utils.SuccessStatusEvent(os.Stdout, "Success! Bhojpur Application runtime is up and running. To get started, go here: https://aka.ms/bhojpur-getting-started")
		}
	},
}

func init() {
	defaultRuntimeVersion := "latest"
	viper.BindEnv("runtime_version_override", "APP_RUNTIME_VERSION")
	runtimeVersionEnv := viper.GetString("runtime_version_override")
	if runtimeVersionEnv != "" {
		defaultRuntimeVersion = runtimeVersionEnv
	}
	defaultDashboardVersion := "latest"
	viper.BindEnv("dashboard_version_override", "APP_DASHBOARD_VERSION")
	dashboardVersionEnv := viper.GetString("dashboard_version_override")
	if dashboardVersionEnv != "" {
		defaultDashboardVersion = dashboardVersionEnv
	}
	InitCmd.Flags().BoolVarP(&kubernetesMode, "kubernetes", "k", false, "Deploy the Bhojpur Application runtime to a Kubernetes cluster")
	InitCmd.Flags().BoolVarP(&wait, "wait", "", false, "Wait for Kubernetes initialization to complete")
	InitCmd.Flags().UintVarP(&timeout, "timeout", "", 300, "The wait timeout for the Kubernetes installation")
	InitCmd.Flags().BoolVarP(&slimMode, "slim", "s", false, "Exclude placement service, Redis, and Zipkin containers from self-hosted installation")
	InitCmd.Flags().StringVarP(&runtimeVersion, "runtime-version", "", defaultRuntimeVersion, "The version of the Bhojpur Application runtime to install, for example: 1.0.0")
	InitCmd.Flags().StringVarP(&dashboardVersion, "dashboard-version", "", defaultDashboardVersion, "The version of the Bhojpur Application dashboard to install, for example: 1.0.0")
	InitCmd.Flags().StringVarP(&initNamespace, "namespace", "n", "app-system", "The Kubernetes namespace to install Bhojpur Application runtime in")
	InitCmd.Flags().BoolVarP(&enableMTLS, "enable-mtls", "", true, "Enable mTLS in your cluster")
	InitCmd.Flags().BoolVarP(&enableHA, "enable-ha", "", false, "Enable high availability (HA) mode")
	InitCmd.Flags().String("network", "", "The Docker network on which to deploy the Bhojpur Application runtime")
	InitCmd.Flags().BoolP("help", "h", false, "Print this help message")
	InitCmd.Flags().StringArrayVar(&values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	InitCmd.Flags().String("image-repository", "", "Custom/Private docker image repository url")
	rootCmd.AddCommand(InitCmd)
}
