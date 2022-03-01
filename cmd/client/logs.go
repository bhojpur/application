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
	"os"

	"github.com/spf13/cobra"

	"github.com/bhojpur/application/pkg/kubernetes"
	"github.com/bhojpur/application/pkg/utils"
)

var (
	logsAppID string
	podName   string
	namespace string
	k8s       bool
)

var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get the runtime sidecar logs for an application. Supported platforms: Kubernetes",
	Example: `
# Get logs of a sample application from target Pod in custom namespace
appctl logs -k --app-id sample --pod-name target --namespace custom
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := kubernetes.Logs(logsAppID, podName, namespace)
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, err.Error())
			os.Exit(1)
		}
		utils.SuccessStatusEvent(os.Stdout, "Fetched logs")
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		kubernetes.CheckForCertExpiry()
	},
}

func init() {
	LogsCmd.Flags().BoolVarP(&k8s, "kubernetes", "k", true, "Get logs from a Kubernetes cluster")
	LogsCmd.Flags().StringVarP(&logsAppID, "app-id", "a", "", "The application id for which logs are needed")
	LogsCmd.Flags().StringVarP(&podName, "pod-name", "p", "", "The name of the pod in Kubernetes, in case your application has multiple pods (optional)")
	LogsCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "The Kubernetes namespace in which your application is deployed")
	LogsCmd.Flags().BoolP("help", "h", false, "Print this help message")
	LogsCmd.MarkFlagRequired("app-id")
	LogsCmd.MarkFlagRequired("kubernetes")
	rootCmd.AddCommand(LogsCmd)
}
