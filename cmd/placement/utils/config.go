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
	RaftID           string
	RaftPeerString   string
	RaftPeers        []raft.PeerInfo
	RaftInMemEnabled bool
	RaftLogStorePath string

	// Placement server configurations
	PlacementPort int
	HealthzPort   int
	CertChainPath string
	TlsEnabled    bool

	ReplicationFactor int

	// Log and metrics configurations
	LoggerOptions   logger.Options
	MetricsExporter metrics.Exporter
}

func NewConfig() *config {
	// Default configuration
	cfg := config{
		RaftID:           "app-placement-0",
		RaftPeerString:   "app-placement-0=127.0.0.1:8201",
		RaftPeers:        []raft.PeerInfo{},
		RaftInMemEnabled: true,
		RaftLogStorePath: "",

		PlacementPort: defaultPlacementPort,
		HealthzPort:   defaultHealthzPort,
		CertChainPath: defaultCredentialsPath,
		TlsEnabled:    false,
	}

	flag.StringVar(&cfg.RaftID, "id", cfg.RaftID, "Placement server ID.")
	flag.StringVar(&cfg.RaftPeerString, "initial-cluster", cfg.RaftPeerString, "raft cluster peers")
	flag.BoolVar(&cfg.RaftInMemEnabled, "inmem-store-enabled", cfg.RaftInMemEnabled, "Enable in-memory log and snapshot store unless --raft-logstore-path is set")
	flag.StringVar(&cfg.RaftLogStorePath, "raft-logstore-path", cfg.RaftLogStorePath, "raft log store path.")
	flag.IntVar(&cfg.PlacementPort, "port", cfg.PlacementPort, "sets the gRPC port for the placement service")
	flag.IntVar(&cfg.HealthzPort, "healthz-port", cfg.HealthzPort, "sets the HTTP port for the healthz server")
	flag.StringVar(&cfg.CertChainPath, "certchain", cfg.CertChainPath, "Path to the credentials directory holding the cert chain")
	flag.BoolVar(&cfg.TlsEnabled, "tls-enabled", cfg.TlsEnabled, "Should TLS be enabled for the placement gRPC server")
	flag.IntVar(&cfg.ReplicationFactor, "replicationFactor", defaultReplicationFactor, "sets the replication factor for actor distribution on vnodes")

	cfg.LoggerOptions = logger.DefaultOptions()
	cfg.LoggerOptions.AttachCmdFlags(flag.StringVar, flag.BoolVar)

	cfg.MetricsExporter = metrics.NewExporter(metrics.DefaultMetricNamespace)
	cfg.MetricsExporter.Options().AttachCmdFlags(flag.StringVar, flag.BoolVar)

	flag.Parse()

	cfg.RaftPeers = parsePeersFromFlag(cfg.RaftPeerString)
	if cfg.RaftLogStorePath != "" {
		cfg.RaftInMemEnabled = false
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
