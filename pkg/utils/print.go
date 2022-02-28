package utils

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
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

const (
	windowsOS = "windows"
)

type Result bool

const (
	Success Result = true
	Failure Result = false
)

var (
	Yellow    = color.New(color.FgHiYellow, color.Bold).SprintFunc()
	Green     = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	Blue      = color.New(color.FgHiBlue, color.Bold).SprintFunc()
	Cyan      = color.New(color.FgCyan, color.Bold, color.Underline).SprintFunc()
	Red       = color.New(color.FgHiRed, color.Bold).Add(color.Italic).SprintFunc()
	White     = color.New(color.FgWhite).SprintFunc()
	WhiteBold = color.New(color.FgWhite, color.Bold).SprintFunc()
)

var logAsJSON bool

func EnableJSONFormat() {
	logAsJSON = true
}

// SuccessStatusEvent reports on a success event.
func SuccessStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if logAsJSON {
		logJSON(w, "success", fmt.Sprintf(fmtstr, a...))
	} else if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "✅  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// FailureStatusEvent reports on a failure event.
func FailureStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if logAsJSON {
		logJSON(w, "failure", fmt.Sprintf(fmtstr, a...))
	} else if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "❌  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// WarningStatusEvent reports on a failure event.
func WarningStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if logAsJSON {
		logJSON(w, "warning", fmt.Sprintf(fmtstr, a...))
	} else if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "⚠  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// PendingStatusEvent reports on a pending event.
func PendingStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if logAsJSON {
		logJSON(w, "pending", fmt.Sprintf(fmtstr, a...))
	} else if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "⌛  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

// InfoStatusEvent reports status information on an event.
func InfoStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	if logAsJSON {
		logJSON(w, "info", fmt.Sprintf(fmtstr, a...))
	} else if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", fmt.Sprintf(fmtstr, a...))
	} else {
		fmt.Fprintf(w, "ℹ️  %s\n", fmt.Sprintf(fmtstr, a...))
	}
}

func Spinner(w io.Writer, fmtstr string, a ...interface{}) func(result Result) {
	msg := fmt.Sprintf(fmtstr, a...)
	var once sync.Once
	var s *spinner.Spinner

	if logAsJSON {
		logJSON(w, "pending", msg)
	} else if runtime.GOOS == windowsOS {
		fmt.Fprintf(w, "%s\n", msg)

		return func(Result) {} // Return a dummy func
	} else {
		s = spinner.New(spinner.CharSets[0], 100*time.Millisecond)
		s.Writer = w
		s.Color("cyan")
		s.Suffix = fmt.Sprintf("  %s", msg)
		s.Start()
	}

	return func(result Result) {
		once.Do(func() {
			if s != nil {
				s.Stop()
			}
			if result {
				SuccessStatusEvent(w, msg)
			} else {
				FailureStatusEvent(w, msg)
			}
		})
	}
}

func logJSON(w io.Writer, status, message string) {
	type jsonLog struct {
		Time    time.Time `json:"time"`
		Status  string    `json:"status"`
		Message string    `json:"msg"`
	}

	l := jsonLog{
		Time:    time.Now().UTC(),
		Status:  status,
		Message: message,
	}
	jsonBytes, err := json.Marshal(&l)
	if err != nil {
		// Fall back on printing the simple message without JSON.
		// This is unlikely.
		fmt.Fprintln(w, message)

		return
	}

	fmt.Fprintf(w, "%s\n", string(jsonBytes))
}
