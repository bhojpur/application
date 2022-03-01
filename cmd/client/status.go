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

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"

	"github.com/bhojpur/application/pkg/kubernetes"
	"github.com/bhojpur/application/pkg/utils"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the health status of runtime services. Supported platforms: Kubernetes",
	Example: `
# Get status of Bhojpur Application services from Kubernetes
appctl status -k 
`,
	Run: func(cmd *cobra.Command, args []string) {
		sc, err := kubernetes.NewStatusClient()
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, err.Error())
			os.Exit(1)
		}
		status, err := sc.Status()
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, err.Error())
			os.Exit(1)
		}
		if len(status) == 0 {
			utils.FailureStatusEvent(os.Stderr, "No status returned. Is the Bhojpur Application initialized in your cluster?")
			os.Exit(1)
		}
		table, err := gocsv.MarshalString(status)
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, err.Error())
			os.Exit(1)
		}

		utils.PrintTable(table)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		kubernetes.CheckForCertExpiry()
	},
}

func init() {
	StatusCmd.Flags().BoolVarP(&k8s, "kubernetes", "k", false, "Show the health status of Bhojpur Application services on Kubernetes cluster")
	StatusCmd.Flags().BoolP("help", "h", false, "Print this help message")
	StatusCmd.MarkFlagRequired("kubernetes")
	rootCmd.AddCommand(StatusCmd)
}
