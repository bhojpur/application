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
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/bhojpur/application/pkg/kubernetes"
	"github.com/bhojpur/application/pkg/utils"
)

var exportPath string

var MTLSCmd = &cobra.Command{
	Use:   "mtls",
	Short: "Check if mTLS is enabled. Supported platforms: Kubernetes",
	Example: `
# Check if mTLS is enabled
appctl mtls -k
`,
	Run: func(cmd *cobra.Command, args []string) {
		enabled, err := kubernetes.IsMTLSEnabled()
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("error checking mTLS: %s", err))
			os.Exit(1)
		}

		status := "disabled"
		if enabled {
			status = "enabled"
		}
		fmt.Printf("Mutual TLS is %s in your Kubernetes cluster \n", status)
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		kubernetes.CheckForCertExpiry()
	},
}

var ExportCMD = &cobra.Command{
	Use:   "export",
	Short: "Export the root CA, issuer cert and key from Kubernetes to local files",
	Example: `
# Export certs to local folder 
appctl mtls export -o ./certs
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := kubernetes.ExportTrustChain(exportPath)
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("error exporting trust chain certs: %s", err))
			os.Exit(1)
		}

		dir, _ := filepath.Abs(exportPath)
		utils.SuccessStatusEvent(os.Stdout, fmt.Sprintf("Trust certs successfully exported to %s", dir))
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		kubernetes.CheckForCertExpiry()
	},
}

var ExpiryCMD = &cobra.Command{
	Use:   "expiry",
	Short: "Checks the expiry of the root certificate",
	Example: `
# Check expiry of Kubernetes certs
appctl mtls expiry
`,
	Run: func(cmd *cobra.Command, args []string) {
		expiry, err := kubernetes.Expiry()
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("error getting root cert expiry: %s", err))
			return
		}

		duration := int(expiry.Sub(time.Now().UTC()).Hours())
		fmt.Printf("Root certificate expires in %v hours. Expiry date: %s", duration, expiry.String())
	},
}

func init() {
	MTLSCmd.Flags().BoolVarP(&kubernetesMode, "kubernetes", "k", false, "Check if mTLS is enabled in a Kubernetes cluster")
	MTLSCmd.Flags().BoolP("help", "h", false, "Print this help message")
	ExportCMD.Flags().StringVarP(&exportPath, "out", "o", ".", "The output directory path to save the certs")
	ExportCMD.Flags().BoolP("help", "h", false, "Print this help message")
	MTLSCmd.MarkFlagRequired("kubernetes")
	MTLSCmd.AddCommand(ExportCMD)
	MTLSCmd.AddCommand(ExpiryCMD)
	rootCmd.AddCommand(MTLSCmd)
}
