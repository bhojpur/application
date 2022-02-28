package main

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
	"flag"
	"time"

	"k8s.io/klog"

	"github.com/bhojpur/service/pkg/utils/logger"

	"github.com/bhojpur/application/pkg/metrics"
	"github.com/bhojpur/application/pkg/operator"
	"github.com/bhojpur/application/pkg/operator/monitoring"
	"github.com/bhojpur/application/pkg/signals"
	"github.com/bhojpur/application/pkg/version"
)

var (
	log                   = logger.NewLogger("app.operator")
	config                string
	certChainPath         string
	disableLeaderElection bool
)

const (
	defaultCredentialsPath = "/var/run/bhojpur/credentials"

	// defaultAppSystemConfigName is the default resource object name for Bhojpur Application System Config.
	defaultAppSystemConfigName = "appsystem"
)

func main() {
	log.Infof("starting Bhojpur Application Operator -- version %s -- commit %s", version.Version(), version.Commit())

	ctx := signals.Context()
	go operator.NewOperator(config, certChainPath, !disableLeaderElection).Run(ctx)
	// The webhooks use their own controller context and stops on SIGTERM and SIGINT.
	go operator.RunWebhooks(!disableLeaderElection)

	<-ctx.Done() // Wait for SIGTERM and SIGINT.

	shutdownDuration := 5 * time.Second
	log.Infof("allowing %s for graceful shutdown to complete", shutdownDuration)
	<-time.After(shutdownDuration)
}

func init() {
	// This resets the flags on klog, which will otherwise try to log to the FS.
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	klogFlags.Set("logtostderr", "true")

	loggerOptions := logger.DefaultOptions()
	loggerOptions.AttachCmdFlags(flag.StringVar, flag.BoolVar)

	metricsExporter := metrics.NewExporter(metrics.DefaultMetricNamespace)
	metricsExporter.Options().AttachCmdFlags(flag.StringVar, flag.BoolVar)

	flag.StringVar(&config, "config", defaultAppSystemConfigName, "Path to config file, or name of a configuration object")
	flag.StringVar(&certChainPath, "certchain", defaultCredentialsPath, "Path to the credentials directory holding the cert chain")

	flag.BoolVar(&disableLeaderElection, "disable-leader-election", false, "Disable leader election for controller manager. ")

	flag.Parse()

	// Apply options to all loggers
	if err := logger.ApplyOptionsToLoggers(&loggerOptions); err != nil {
		log.Fatal(err)
	} else {
		log.Infof("log level set to: %s", loggerOptions.OutputLevel)
	}

	// Initialize Bhojpur Application runtime metrics exporter
	if err := metricsExporter.Init(); err != nil {
		log.Fatal(err)
	}

	if err := monitoring.InitMetrics(); err != nil {
		log.Fatal(err)
	}
}
