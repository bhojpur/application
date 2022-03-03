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
	"os/signal"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/bhojpur/application/pkg/kubernetes"
	"github.com/bhojpur/application/pkg/standalone"
	"github.com/bhojpur/application/pkg/utils"
)

const (
	// dashboardSvc is the name of the Bhojpur Dashboard service running in cluster.
	dashboardSvc = "app-dashboard"

	// defaultHost is the default host used for port forwarding for `appctl dashboard`.
	defaultHost = "localhost"

	// defaultLocalPort is the default local port used for port forwarding for `appctl dashboard`.
	defaultLocalPort = 8000

	// appSystemNamespace is the namespace "app-system" (recommended Bhojpur Application runtime install namespace).
	appSystemNamespace = "app-system"

	// defaultNamespace is the default namespace (appctl init -k installation).
	defaultNamespace = "default"

	// remotePort is the port Bhojpur Application dashboard pod is listening on.
	remotePort = 8080
)

var (
	dashboardNamespace  string
	dashboardHost       string
	dashboardLocalPort  int
	dashboardVersionCmd bool
)

var DashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Start the dashboard. Supported platforms: Kubernetes and self-hosted",
	Example: `
# Start dashboard locally
appctl dashboard

# Start dashboard locally in a specified port 
appctl dashboard -p 9999

# Port forward to dashboard in Kubernetes 
appctl dashboard -k 

# Port forward to dashboard in Kubernetes on all addresses in a specified port
appctl dashboard -k -p 9999 -a 0.0.0.0

# Port forward to dashboard in Kubernetes using a port
appctl dashboard -k -p 9999
`,
	Run: func(cmd *cobra.Command, args []string) {
		if dashboardVersionCmd {
			fmt.Println(standalone.GetDashboardVersion())
			os.Exit(0)
		}

		if !utils.IsAddressLegal(dashboardHost) {
			utils.FailureStatusEvent(os.Stdout, "Invalid address: %s", dashboardHost)
			os.Exit(1)
		}

		if dashboardLocalPort <= 0 {
			utils.FailureStatusEvent(os.Stderr, "Invalid port: %v", dashboardLocalPort)
			os.Exit(1)
		}

		if kubernetesMode {
			config, client, err := kubernetes.GetKubeConfigClient()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, "Failed to initialize kubernetes client: %s", err.Error())
				os.Exit(1)
			}

			// search for Bhojpur Dashboard service namespace in order:
			// user-supplied namespace, app-system, default
			namespaces := []string{dashboardNamespace}
			if dashboardNamespace != appSystemNamespace {
				namespaces = append(namespaces, appSystemNamespace)
			}
			if dashboardNamespace != defaultNamespace {
				namespaces = append(namespaces, defaultNamespace)
			}

			foundNamespace := ""
			for _, namespace := range namespaces {
				ok, _ := kubernetes.CheckPodExists(client, namespace, nil, dashboardSvc)
				if ok {
					foundNamespace = namespace
					break
				}
			}

			// if the service is not found, try to search all pods
			if foundNamespace == "" {
				ok, nspace := kubernetes.CheckPodExists(client, "", nil, dashboardSvc)

				// if the service is found, tell the user to try with the found namespace
				// if the service is still not found, throw an error
				if ok {
					utils.InfoStatusEvent(os.Stdout, "Bhojpur Dashboard found in namespace: %s. Run appctl dashboard -k -n %s to use this namespace.", nspace, nspace)
				} else {
					utils.FailureStatusEvent(os.Stderr, "Failed to find Bhojpur Dashboard in cluster. Check status of appctl dashboard in the cluster.")
				}
				os.Exit(1)
			}

			// manage termination of port forwarding connection on interrupt
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)
			defer signal.Stop(signals)

			portForward, err := kubernetes.NewPortForward(
				config,
				foundNamespace,
				dashboardSvc,
				dashboardHost,
				dashboardLocalPort,
				remotePort,
				false,
			)
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, "%s\n", err)
				os.Exit(1)
			}

			// initialize port forwarding
			if err = portForward.Init(); err != nil {
				utils.FailureStatusEvent(os.Stderr, "Error in port forwarding: %s\nCheck for `appctl dashboard` running in other terminal sessions, or use the `--port` flag to use a different port.\n", err)
				os.Exit(1)
			}

			// block until interrupt signal is received
			go func() {
				<-signals
				portForward.Stop()
			}()

			// url for dashboard after port forwarding
			var webURL string = fmt.Sprintf("http://%s:%d", dashboardHost, dashboardLocalPort)

			utils.InfoStatusEvent(os.Stdout, fmt.Sprintf("Bhojpur Dashboard found in namespace:\t%s", foundNamespace))
			utils.InfoStatusEvent(os.Stdout, fmt.Sprintf("Bhojpur Dashboard available at:\t%s\n", webURL))

			err = browser.OpenURL(webURL)
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, "Failed to start Bhojpur Dashboard in browser automatically")
				utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Visit %s in your browser to view the Bhojpur Dashboard", webURL))
			}

			<-portForward.GetStop()
		} else {
			// Standalone mode
			err := standalone.NewDashboardCmd(dashboardLocalPort).Run()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, "Bhojpur Dashboard not found. Is the Bhojpur Application runtime engine installed?")
			}
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if kubernetesMode {
			kubernetes.CheckForCertExpiry()
		}
	},
}

func init() {
	DashboardCmd.Flags().BoolVarP(&kubernetesMode, "kubernetes", "k", false, "Opens the Bhojpur Dashboard in local browser via local proxy to Kubernetes cluster")
	DashboardCmd.Flags().BoolVarP(&dashboardVersionCmd, "version", "v", false, "Print the version for Bhojpur Dashboard")
	DashboardCmd.Flags().StringVarP(&dashboardHost, "address", "a", defaultHost, "Address to listen on. Only accepts IP address or localhost as a value")
	DashboardCmd.Flags().IntVarP(&dashboardLocalPort, "port", "p", defaultLocalPort, "The local port on which to serve Bhojpur Dashboard")
	DashboardCmd.Flags().StringVarP(&dashboardNamespace, "namespace", "n", appSystemNamespace, "The namespace where Bhojpur Dashboard is running")
	DashboardCmd.Flags().BoolP("help", "h", false, "Print this help message")
	rootCmd.AddCommand(DashboardCmd)
}
