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
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/bhojpur/application/pkg/standalone"
	"github.com/bhojpur/application/pkg/utils"
)

const defaultHTTPVerb = http.MethodPost

var (
	invokeAppID     string
	invokeAppMethod string
	invokeData      string
	invokeVerb      string
	invokeDataFile  string
	invokeSocket    string
)

var InvokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Invoke a method on a given Bhojpur Application. Supported platforms: Self-hosted",
	Example: `
# Invoke a sample method on target application with POST Verb
appctl invoke --app-id target --method sample --data '{"key":"value"}

# Invoke a sample method on target app with GET Verb
appctl invoke --app-id target --method sample --verb GET

# Invoke a sample method on target app with GET Verb using Unix domain socket
appctl invoke --unix-domain-socket --app-id target --method sample --verb GET
`,
	Run: func(cmd *cobra.Command, args []string) {
		bytePayload := []byte{}
		var err error
		if invokeDataFile != "" && invokeData != "" {
			utils.FailureStatusEvent(os.Stderr, "Only one of --data and --data-file allowed in the same invoke command")
			os.Exit(1)
		}

		if invokeDataFile != "" {
			bytePayload, err = ioutil.ReadFile(invokeDataFile)
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, "Error reading payload from '%s'. Error: %s", invokeDataFile, err)
				os.Exit(1)
			}
		} else if invokeData != "" {
			bytePayload = []byte(invokeData)
		}
		client := standalone.NewClient()

		// TODO: add Windows support
		if invokeSocket != "" {
			if runtime.GOOS == "windows" {
				utils.FailureStatusEvent(os.Stderr, "The unix-domain-socket option is not supported on Windows")
				os.Exit(1)
			} else {
				utils.WarningStatusEvent(os.Stdout, "Unix domain sockets are currently a preview feature")
			}
		}

		response, err := client.Invoke(invokeAppID, invokeAppMethod, bytePayload, invokeVerb, invokeSocket)
		if err != nil {
			err = fmt.Errorf("error invoking Bhojpur Application %s: %s", invokeAppID, err)
			utils.FailureStatusEvent(os.Stderr, err.Error())
			return
		}

		if response != "" {
			fmt.Println(response)
		}
		utils.SuccessStatusEvent(os.Stdout, "Bhojpur Application invoked successfully")
	},
}

func init() {
	InvokeCmd.Flags().StringVarP(&invokeAppID, "app-id", "a", "", "The application id to invoke")
	InvokeCmd.Flags().StringVarP(&invokeAppMethod, "method", "m", "", "The method to invoke")
	InvokeCmd.Flags().StringVarP(&invokeData, "data", "d", "", "The JSON serialized data string (optional)")
	InvokeCmd.Flags().StringVarP(&invokeVerb, "verb", "v", defaultHTTPVerb, "The HTTP verb to use")
	InvokeCmd.Flags().StringVarP(&invokeDataFile, "data-file", "f", "", "A file containing the JSON serialized data (optional)")
	InvokeCmd.Flags().BoolP("help", "h", false, "Print this help message")
	InvokeCmd.Flags().StringVarP(&invokeSocket, "unix-domain-socket", "u", "", "Path to a unix domain socket dir. If specified, Bhojpur Application API servers will use Unix Domain Sockets")
	InvokeCmd.MarkFlagRequired("app-id")
	InvokeCmd.MarkFlagRequired("method")
	rootCmd.AddCommand(InvokeCmd)
}
