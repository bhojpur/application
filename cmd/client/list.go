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

	"github.com/gocarina/gocsv"
	"github.com/spf13/cobra"

	"github.com/bhojpur/application/pkg/kubernetes"
	"github.com/bhojpur/application/pkg/standalone"
	"github.com/bhojpur/application/pkg/utils"
)

var outputFormat string

func outputList(list interface{}, length int) {
	if outputFormat == "json" || outputFormat == "yaml" {
		err := utils.PrintDetail(os.Stdout, outputFormat, list)
		if err != nil {
			utils.FailureStatusEvent(os.Stdout, err.Error())
			os.Exit(1)
		}
	} else {
		table, err := gocsv.MarshalString(list)
		if err != nil {
			utils.FailureStatusEvent(os.Stdout, err.Error())
			os.Exit(1)
		}

		// Standalone mode displays a separate message when no instances are found.
		if !kubernetesMode && length == 0 {
			fmt.Println("No Bhojpur Application runtime instances found.")
			return
		}

		utils.PrintTable(table)
	}
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Bhojpur Application instances. Supported platforms: Kubernetes and self-hosted",
	Example: `
# List the Bhojpur Application instances in self-hosted mode
appctl list

# List the Bhojpur Application instances in Kubernetes mode
appctl list -k
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if outputFormat != "" && outputFormat != "json" && outputFormat != "yaml" && outputFormat != "table" {
			utils.FailureStatusEvent(os.Stdout, "An invalid output format was specified.")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if kubernetesMode {
			list, err := kubernetes.List()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, err.Error())
				os.Exit(1)
			}

			outputList(list, len(list))
		} else {
			list, err := standalone.List()
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, err.Error())
				os.Exit(1)
			}

			outputList(list, len(list))
		}
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if kubernetesMode {
			kubernetes.CheckForCertExpiry()
		}
	},
}

func init() {
	ListCmd.Flags().BoolVarP(&kubernetesMode, "kubernetes", "k", false, "List all Bhojpur Application runtime pods in a Kubernetes cluster")
	ListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "The output format of the list. Valid values are: json, yaml, or table (default)")
	ListCmd.Flags().BoolP("help", "h", false, "Print this help message")
	rootCmd.AddCommand(ListCmd)
}
