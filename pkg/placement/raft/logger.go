package raft

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
	"io"
	"log"

	"github.com/hashicorp/go-hclog"

	"github.com/bhojpur/service/pkg/utils/logger"
)

var logging = logger.NewLogger("app.placement.raft")

func newLoggerAdapter() hclog.Logger {
	return &loggerAdapter{}
}

// loggerAdapter is the adapter to integrate with Bhojpur Application runtime logger.
type loggerAdapter struct{}

func (l *loggerAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	switch level {
	case hclog.Debug:
		logging.Debugf(msg, args)
	case hclog.Warn:
		logging.Debugf(msg, args)
	case hclog.Error:
		logging.Debugf(msg, args)
	default:
		logging.Debugf(msg, args)
	}
}

func (l *loggerAdapter) Trace(msg string, args ...interface{}) {
	logging.Debugf(msg, args)
}

func (l *loggerAdapter) Debug(msg string, args ...interface{}) {
	logging.Debugf(msg, args)
}

func (l *loggerAdapter) Info(msg string, args ...interface{}) {
	logging.Debugf(msg, args)
}

func (l *loggerAdapter) Warn(msg string, args ...interface{}) {
	logging.Debugf(msg, args)
}

func (l *loggerAdapter) Error(msg string, args ...interface{}) {
	logging.Debugf(msg, args)
}

func (l *loggerAdapter) IsTrace() bool { return false }

func (l *loggerAdapter) IsDebug() bool { return true }

func (l *loggerAdapter) IsInfo() bool { return false }

func (l *loggerAdapter) IsWarn() bool { return false }

func (l *loggerAdapter) IsError() bool { return false }

func (l *loggerAdapter) ImpliedArgs() []interface{} { return []interface{}{} }

func (l *loggerAdapter) With(args ...interface{}) hclog.Logger { return l }

func (l *loggerAdapter) Name() string { return "app" }

func (l *loggerAdapter) Named(name string) hclog.Logger { return l }

func (l *loggerAdapter) ResetNamed(name string) hclog.Logger { return l }

func (l *loggerAdapter) SetLevel(level hclog.Level) {}

func (l *loggerAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(l.StandardWriter(opts), "placement-raft", log.LstdFlags)
}

func (l *loggerAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return io.Discard
}
