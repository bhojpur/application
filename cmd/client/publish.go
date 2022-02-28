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
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/bhojpur/application/pkg/standalone"
	"github.com/bhojpur/application/pkg/utils"
)

var (
	publishAppID       string
	pubsubName         string
	publishTopic       string
	publishPayload     string
	publishPayloadFile string
	publishSocket      string
)

var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a pub-sub event. Supported platforms: Self-hosted",
	Example: `
# Publish to a sample Topic in target pubsub via a publishing application
appctl publish --publish-app-id myapp --pubsub target --topic sample --data '{"key":"value"}'

# Publish to a sample Topic in target pubsub via a publishing application using Unix domain socket
appctl publish --enable-domain-socket --publish-app-id myapp --pubsub target --topic sample --data '{"key":"value"}'
`,
	Run: func(cmd *cobra.Command, args []string) {
		bytePayload := []byte{}
		var err error
		if publishPayloadFile != "" && publishPayload != "" {
			utils.FailureStatusEvent(os.Stderr, "Only one of --data and --data-file allowed in the same publish command")
			os.Exit(1)
		}

		if publishPayloadFile != "" {
			bytePayload, err = ioutil.ReadFile(publishPayloadFile)
			if err != nil {
				utils.FailureStatusEvent(os.Stderr, "Error reading payload from '%s'. Error: %s", publishPayloadFile, err)
				os.Exit(1)
			}
		} else if publishPayload != "" {
			bytePayload = []byte(publishPayload)
		}

		client := standalone.NewClient()
		// TODO: add Windows support
		if publishSocket != "" {
			if runtime.GOOS == "windows" {
				utils.FailureStatusEvent(os.Stderr, "The unix-domain-socket option is not supported on Windows")
				os.Exit(1)
			} else {
				utils.WarningStatusEvent(os.Stdout, "Unix domain sockets are currently a preview feature")
			}
		}
		err = client.Publish(publishAppID, pubsubName, publishTopic, bytePayload, publishSocket)
		if err != nil {
			utils.FailureStatusEvent(os.Stderr, fmt.Sprintf("Error publishing topic %s: %s", publishTopic, err))
			os.Exit(1)
		}

		utils.SuccessStatusEvent(os.Stdout, "Event published successfully")
	},
}

func init() {
	PublishCmd.Flags().StringVarP(&publishAppID, "publish-app-id", "i", "", "The ID of the publishing application")
	PublishCmd.Flags().StringVarP(&pubsubName, "pubsub", "p", "", "The name of the pub/sub component")
	PublishCmd.Flags().StringVarP(&publishTopic, "topic", "t", "", "The topic to be published to")
	PublishCmd.Flags().StringVarP(&publishPayload, "data", "d", "", "The JSON serialized data string (optional)")
	PublishCmd.Flags().StringVarP(&publishPayloadFile, "data-file", "f", "", "A file containing the JSON serialized data (optional)")
	PublishCmd.Flags().StringVarP(&publishSocket, "unix-domain-socket", "u", "", "Path to a unix domain socket dir. If specified, Bhojpur Application API servers will use Unix Domain Sockets")
	PublishCmd.Flags().BoolP("help", "h", false, "Print this help message")
	PublishCmd.MarkFlagRequired("publish-app-id")
	PublishCmd.MarkFlagRequired("topic")
	PublishCmd.MarkFlagRequired("pubsub")
	rootCmd.AddCommand(PublishCmd)
}
