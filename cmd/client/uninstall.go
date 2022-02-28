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
	uninstallNamespace  string
	uninstallKubernetes bool
	uninstallAll        bool
)

// UninstallCmd is a command from removing a Bhojpur Application installation.
var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Bhojpur Application runtime. Supported platforms: Kubernetes and self-hosted",
	Example: `
# Uninstall from self-hosted mode
appctl uninstall

# Uninstall from self-hosted mode and remove .bhojpur directory, Redis, Placement and Zipkin containers
appctl uninstall --all

# Uninstall from Kubernetes
appctl uninstall -k
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("network", cmd.Flags().Lookup("network"))
		viper.BindPFlag("install-path", cmd.Flags().Lookup("install-path"))
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		if uninstallKubernetes {
			utils.InfoStatusEvent(os.Stdout, "Removing Bhojpur Application from your cluster...")
			err = kubernetes.Uninstall(uninstallNamespace, uninstallAll, timeout)
		} else {
			utils.InfoStatusEvent(os.Stdout, "Removing Bhojpur Application from your machine...")
			dockerNetwork := viper.GetString("network")
			err = standalone.Uninstall(uninstallAll, dockerNetwork)
		}

		if err != nil {
			utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Error removing Bhojpur Application: %s", err))
		} else {
			utils.SuccessStatusEvent(os.Stdout, "Bhojpu Rapplication has been removed successfully")
		}
	},
}

func init() {
	UninstallCmd.Flags().BoolVarP(&uninstallKubernetes, "kubernetes", "k", false, "Uninstall Bhojpur Application from a Kubernetes cluster")
	UninstallCmd.Flags().UintVarP(&timeout, "timeout", "", 300, "The timeout for the Kubernetes uninstall")
	UninstallCmd.Flags().BoolVar(&uninstallAll, "all", false, "Remove .bhojpur directory, Redis, Placement and Zipkin containers on local machine, and CRDs on a Kubernetes cluster")
	UninstallCmd.Flags().String("network", "", "The Docker network from which to remove the Bhojpur Application runtime")
	UninstallCmd.Flags().StringVarP(&uninstallNamespace, "namespace", "n", "app-system", "The Kubernetes namespace to uninstall Bhojpur Application runtime from")
	UninstallCmd.Flags().BoolP("help", "h", false, "Print this help message")
	rootCmd.AddCommand(UninstallCmd)
}
