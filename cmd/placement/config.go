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
	"strings"

	"github.com/bhojpur/service/pkg/utils/logger"

	"github.com/bhojpur/application/pkg/metrics"
	"github.com/bhojpur/application/pkg/placement/raft"
)

const (
	defaultCredentialsPath   = "/var/run/bhojpur/credentials"
	defaultHealthzPort       = 8080
	defaultPlacementPort     = 50005
	defaultReplicationFactor = 100
)

type config struct {
	// Raft protocol configurations
	raftID           string
	raftPeerString   string
	raftPeers        []raft.PeerInfo
	raftInMemEnabled bool
	raftLogStorePath string

	// Placement server configurations
	placementPort int
	healthzPort   int
	certChainPath string
	tlsEnabled    bool

	replicationFactor int

	// Log and metrics configurations
	loggerOptions   logger.Options
	metricsExporter metrics.Exporter
}

func newConfig() *config {
	// Default configuration
	cfg := config{
		raftID:           "app-placement-0",
		raftPeerString:   "app-placement-0=127.0.0.1:8201",
		raftPeers:        []raft.PeerInfo{},
		raftInMemEnabled: true,
		raftLogStorePath: "",

		placementPort: defaultPlacementPort,
		healthzPort:   defaultHealthzPort,
		certChainPath: defaultCredentialsPath,
		tlsEnabled:    false,
	}

	flag.StringVar(&cfg.raftID, "id", cfg.raftID, "Placement server ID.")
	flag.StringVar(&cfg.raftPeerString, "initial-cluster", cfg.raftPeerString, "raft cluster peers")
	flag.BoolVar(&cfg.raftInMemEnabled, "inmem-store-enabled", cfg.raftInMemEnabled, "Enable in-memory log and snapshot store unless --raft-logstore-path is set")
	flag.StringVar(&cfg.raftLogStorePath, "raft-logstore-path", cfg.raftLogStorePath, "raft log store path.")
	flag.IntVar(&cfg.placementPort, "port", cfg.placementPort, "sets the gRPC port for the placement service")
	flag.IntVar(&cfg.healthzPort, "healthz-port", cfg.healthzPort, "sets the HTTP port for the healthz server")
	flag.StringVar(&cfg.certChainPath, "certchain", cfg.certChainPath, "Path to the credentials directory holding the cert chain")
	flag.BoolVar(&cfg.tlsEnabled, "tls-enabled", cfg.tlsEnabled, "Should TLS be enabled for the placement gRPC server")
	flag.IntVar(&cfg.replicationFactor, "replicationFactor", defaultReplicationFactor, "sets the replication factor for actor distribution on vnodes")

	cfg.loggerOptions = logger.DefaultOptions()
	cfg.loggerOptions.AttachCmdFlags(flag.StringVar, flag.BoolVar)

	cfg.metricsExporter = metrics.NewExporter(metrics.DefaultMetricNamespace)
	cfg.metricsExporter.Options().AttachCmdFlags(flag.StringVar, flag.BoolVar)

	flag.Parse()

	cfg.raftPeers = parsePeersFromFlag(cfg.raftPeerString)
	if cfg.raftLogStorePath != "" {
		cfg.raftInMemEnabled = false
	}

	return &cfg
}

func parsePeersFromFlag(val string) []raft.PeerInfo {
	peers := []raft.PeerInfo{}

	p := strings.Split(val, ",")
	for _, addr := range p {
		peer := strings.Split(addr, "=")
		if len(peer) != 2 {
			continue
		}

		peers = append(peers, raft.PeerInfo{
			ID:      strings.TrimSpace(peer[0]),
			Address: strings.TrimSpace(peer[1]),
		})
	}

	return peers
}
