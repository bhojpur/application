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

	"github.com/bhojpur/application/pkg/standalone"
	"github.com/bhojpur/application/pkg/utils"
)

var stopAppID string

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Bhojpur Application instances and their associated apps. Supported platforms: Self-hosted",
	Example: `
# Stop the Bhojpur Application
appctl stop --app-id <ID>
`,
	Run: func(cmd *cobra.Command, args []string) {
		if stopAppID != "" {
			args = append(args, stopAppID)
		}
		for _, appID := range args {
			err := standalone.Stop(appID)
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, "failed to stop app id %s: %s", appID, err)
			} else {
				utils.SuccessStatusEvent(os.Stdout, "app stopped successfully: %s", appID)
			}
		}
	},
}

func init() {
	StopCmd.Flags().StringVarP(&stopAppID, "app-id", "a", "", "The application id to be stopped")
	StopCmd.Flags().BoolP("help", "h", false, "Print this help message")
	rootCmd.AddCommand(StopCmd)
}
