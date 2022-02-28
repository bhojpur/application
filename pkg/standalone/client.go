package standalone

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

type AppProcess interface {
	List() ([]ListOutput, error)
}

type appProcess struct {
}

// Client is the interface the wraps all the methods exposed by the Bhojpur Application CLI.
type Client interface {
	// Invoke is a command to invoke a remote or local Bhojpur Application instance
	Invoke(appID, method string, data []byte, verb string, socket string) (string, error)
	// Publish is used to publish event to a topic in a pubsub for an app ID.
	Publish(publishAppID, pubsubName, topic string, payload []byte, socket string) error
}

type Standalone struct {
	process AppProcess
}

func NewClient() Client {
	return &Standalone{process: &appProcess{}}
}
