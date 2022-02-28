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

var upgradeRuntimeVersion string

var UpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades or downgrades a Bhojpur Application runtime control plane installation in a cluster. Supported platforms: Kubernetes",
	Example: `
# Upgrade the Bhojpur Application runtime in Kubernetes
appctl upgrade -k

# See more at: https://docs.bhojpur.net/getting-started/
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := kubernetes.Upgrade(kubernetes.UpgradeConfig{
			RuntimeVersion: upgradeRuntimeVersion,
			Args:           values,
			Timeout:        timeout,
		})
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, "Failed to upgrade Bhojpur Application runtime: %s", err)
			os.Exit(1)
		}
		utils.SuccessStatusEvent(os.Stdout, "Bhojpur Application runtime control plane successfully upgraded to version %s. Make sure your deployments are restarted to pick up the latest sidecar version.", upgradeRuntimeVersion)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		kubernetes.CheckForCertExpiry()
	},
}

func init() {
	UpgradeCmd.Flags().BoolVarP(&kubernetesMode, "kubernetes", "k", false, "Upgrade or downgrade Bhojpur Application runtime in a Kubernetes cluster")
	UpgradeCmd.Flags().UintVarP(&timeout, "timeout", "", 300, "The timeout for the Kubernetes upgrade")
	UpgradeCmd.Flags().StringVarP(&upgradeRuntimeVersion, "runtime-version", "", "", "The version of the Bhojpur Application runtime to upgrade or downgrade to, for example: 1.0.0")
	UpgradeCmd.Flags().BoolP("help", "h", false, "Print this help message")
	UpgradeCmd.Flags().StringArrayVar(&values, "set", []string{}, "set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")

	UpgradeCmd.MarkFlagRequired("runtime-version")
	UpgradeCmd.MarkFlagRequired("kubernetes")

	rootCmd.AddCommand(UpgradeCmd)
}
