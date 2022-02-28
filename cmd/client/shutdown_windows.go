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
	"os/signal"
	"syscall"

	"github.com/bhojpur/application/pkg/utils"
	"golang.org/x/sys/windows"
)

func setupShutdownNotify(sigCh chan os.Signal) {
	//This will catch Ctrl-C
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// Unlike Linux/Mac, you can't just send a SIGTERM from another process
	// In order for 'appctl stop' to be able to signal gracefully we must use a named event in Windows
	go func() {
		eventName, _ := syscall.UTF16FromString(fmt.Sprintf("app_cli_%v", os.Getpid()))
		eventHandle, _ := windows.CreateEvent(nil, 0, 0, &eventName[0])
		_, err := windows.WaitForSingleObject(eventHandle, windows.INFINITE)
		if err != nil {
			utils.WarningStatusEvent(os.Stdout, "Unable to wait for shutdown event. 'appctl stop' will not work. Error: %s", err.Error())
			return
		}
		sigCh <- os.Interrupt
	}()
}
