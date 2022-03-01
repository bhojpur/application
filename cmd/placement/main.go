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
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bhojpur/service/pkg/utils/logger"

	utils "github.com/bhojpur/application/cmd/placement/utils"
	"github.com/bhojpur/application/pkg/credentials"
	"github.com/bhojpur/application/pkg/fswatcher"
	"github.com/bhojpur/application/pkg/health"
	"github.com/bhojpur/application/pkg/placement"
	"github.com/bhojpur/application/pkg/placement/hashing"
	"github.com/bhojpur/application/pkg/placement/monitoring"
	"github.com/bhojpur/application/pkg/placement/raft"
	"github.com/bhojpur/application/pkg/version"
)

var log = logger.NewLogger("app.placement")

const gracefulTimeout = 10 * time.Second

func main() {
	log.Infof("starting Bhojpur Application Placement Service -- version %s -- commit %s", version.Version(), version.Commit())

	cfg := utils.NewConfig()

	// Apply options to all loggers.
	if err := logger.ApplyOptionsToLoggers(&cfg.LoggerOptions); err != nil {
		log.Fatal(err)
	}
	log.Infof("log level set to: %s", cfg.LoggerOptions.OutputLevel)

	// Initialize Bhojpur Application runtime metrics for placement.
	if err := cfg.MetricsExporter.Init(); err != nil {
		log.Fatal(err)
	}

	if err := monitoring.InitMetrics(); err != nil {
		log.Fatal(err)
	}

	// Start Raft cluster.
	raftServer := raft.New(cfg.RaftID, cfg.RaftInMemEnabled, cfg.RaftPeers, cfg.RaftLogStorePath)
	if raftServer == nil {
		log.Fatal("failed to create raft server.")
	}

	if err := raftServer.StartRaft(nil); err != nil {
		log.Fatalf("failed to start Raft Server: %v", err)
	}

	// Start Placement gRPC server.
	hashing.SetReplicationFactor(cfg.ReplicationFactor)
	apiServer := placement.NewPlacementService(raftServer)
	var certChain *credentials.CertChain
	if cfg.TlsEnabled {
		certChain = loadCertChains(cfg.CertChainPath)
	}

	go apiServer.MonitorLeadership()
	go apiServer.Run(strconv.Itoa(cfg.PlacementPort), certChain)
	log.Infof("Bhojpur Application Placement server started on port %d", cfg.PlacementPort)

	// Start Healthz endpoint.
	go startHealthzServer(cfg.HealthzPort)

	// Relay incoming process signal to exit placement gracefully
	signalCh := make(chan os.Signal, 10)
	gracefulExitCh := make(chan struct{})
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(signalCh)

	<-signalCh

	// Shutdown servers
	go func() {
		apiServer.Shutdown()
		raftServer.Shutdown()
		close(gracefulExitCh)
	}()

	select {
	case <-time.After(gracefulTimeout):
		log.Info("Timeout on graceful leave. Exiting...")
		os.Exit(1)

	case <-gracefulExitCh:
		log.Info("Gracefully exit.")
		os.Exit(0)
	}
}

func startHealthzServer(healthzPort int) {
	healthzServer := health.NewServer(log)
	healthzServer.Ready()

	if err := healthzServer.Run(context.Background(), healthzPort); err != nil {
		log.Fatalf("failed to start healthz server: %s", err)
	}
}

func loadCertChains(certChainPath string) *credentials.CertChain {
	tlsCreds := credentials.NewTLSCredentials(certChainPath)

	log.Info("mTLS enabled, getting tls certificates")
	// try to load certs from disk, if not yet there, start a watch on the local filesystem
	chain, err := credentials.LoadFromDisk(tlsCreds.RootCertPath(), tlsCreds.CertPath(), tlsCreds.KeyPath())
	if err != nil {
		fsevent := make(chan struct{})

		go func() {
			log.Infof("starting watch for certs on filesystem: %s", certChainPath)
			err = fswatcher.Watch(context.Background(), tlsCreds.Path(), fsevent)
			if err != nil {
				log.Fatal("error starting watch on filesystem: %s", err)
			}
		}()

		<-fsevent
		log.Info("certificates detected")

		chain, err = credentials.LoadFromDisk(tlsCreds.RootCertPath(), tlsCreds.CertPath(), tlsCreds.KeyPath())
		if err != nil {
			log.Fatal("failed to load cert chain from disk: %s", err)
		}
	}

	log.Info("tls certificates loaded successfully")

	return chain
}
