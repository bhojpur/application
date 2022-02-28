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
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bhojpur/service/pkg/utils/logger"

	"github.com/bhojpur/application/pkg/credentials"
	"github.com/bhojpur/application/pkg/fswatcher"
	"github.com/bhojpur/application/pkg/health"
	"github.com/bhojpur/application/pkg/metrics"
	"github.com/bhojpur/application/pkg/sentry"
	"github.com/bhojpur/application/pkg/sentry/config"
	"github.com/bhojpur/application/pkg/sentry/monitoring"
	"github.com/bhojpur/application/pkg/signals"
	"github.com/bhojpur/application/pkg/version"
)

var log = logger.NewLogger("app.sentry")

const (
	defaultCredentialsPath = "/var/run/bhojpur/credentials"
	// defaultAppSystemConfigName is the default resource object name for Bhojpur Application System Config.
	defaultAppSystemConfigName = "appsystem"

	healthzPort = 8080
)

func main() {
	configName := flag.String("config", defaultAppSystemConfigName, "Path to config file, or name of a configuration object")
	credsPath := flag.String("issuer-credentials", defaultCredentialsPath, "Path to the credentials directory holding the issuer data")
	trustDomain := flag.String("trust-domain", "localhost", "The CA trust domain")

	loggerOptions := logger.DefaultOptions()
	loggerOptions.AttachCmdFlags(flag.StringVar, flag.BoolVar)

	metricsExporter := metrics.NewExporter(metrics.DefaultMetricNamespace)
	metricsExporter.Options().AttachCmdFlags(flag.StringVar, flag.BoolVar)

	flag.Parse()

	// Apply options to all loggers
	if err := logger.ApplyOptionsToLoggers(&loggerOptions); err != nil {
		log.Fatal(err)
	}

	log.Infof("starting Bhojpur Application Sentry Certificate Authority -- version %s -- commit %s", version.Version(), version.Commit())
	log.Infof("log level set to: %s", loggerOptions.OutputLevel)

	// Initialize Bhojpur Application runtime metrics exporter
	if err := metricsExporter.Init(); err != nil {
		log.Fatal(err)
	}

	if err := monitoring.InitMetrics(); err != nil {
		log.Fatal(err)
	}

	issuerCertPath := filepath.Join(*credsPath, credentials.IssuerCertFilename)
	issuerKeyPath := filepath.Join(*credsPath, credentials.IssuerKeyFilename)
	rootCertPath := filepath.Join(*credsPath, credentials.RootCertFilename)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx := signals.Context()
	config, err := config.FromConfigName(*configName)
	if err != nil {
		log.Warn(err)
	}
	config.IssuerCertPath = issuerCertPath
	config.IssuerKeyPath = issuerKeyPath
	config.RootCertPath = rootCertPath
	config.TrustDomain = *trustDomain

	watchDir := filepath.Dir(config.IssuerCertPath)

	ca := sentry.NewSentryCA()

	log.Infof("starting watch on filesystem directory: %s", watchDir)
	issuerEvent := make(chan struct{})
	ready := make(chan bool)

	go ca.Run(ctx, config, ready)

	<-ready

	go fswatcher.Watch(ctx, watchDir, issuerEvent)

	go func() {
		for range issuerEvent {
			monitoring.IssuerCertChanged()
			log.Warn("issuer credentials changed. reloading")
			ca.Restart(ctx, config)
		}
	}()

	go func() {
		healthzServer := health.NewServer(log)
		healthzServer.Ready()

		err := healthzServer.Run(ctx, healthzPort)
		if err != nil {
			log.Fatalf("failed to start healthz server: %s", err)
		}
	}()

	<-stop
	shutdownDuration := 5 * time.Second
	log.Infof("allowing %s for graceful shutdown to complete", shutdownDuration)
	<-time.After(shutdownDuration)
}
