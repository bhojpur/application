package grpc

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
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pkg/errors"
	grpc_go "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/bhojpur/service/pkg/utils/logger"

	internalv1pb "github.com/bhojpur/application/pkg/api/v1/internals"
	runtimev1pb "github.com/bhojpur/application/pkg/api/v1/runtime"
	"github.com/bhojpur/application/pkg/config"
	diag "github.com/bhojpur/application/pkg/diagnostics"
	diag_utils "github.com/bhojpur/application/pkg/diagnostics/utils"
	"github.com/bhojpur/application/pkg/messaging"
	auth "github.com/bhojpur/application/pkg/runtime/security"
)

const (
	certWatchInterval              = time.Second * 3
	renewWhenPercentagePassed      = 70
	apiServer                      = "apiServer"
	internalServer                 = "internalServer"
	defaultMaxConnectionAgeSeconds = 30
)

// Server is an interface for the Bhojpur Application runtime gRPC server.
type Server interface {
	io.Closer
	StartNonBlocking() error
}

type server struct {
	api                API
	config             ServerConfig
	tracingSpec        config.TracingSpec
	metricSpec         config.MetricSpec
	authenticator      auth.Authenticator
	servers            []*grpc_go.Server
	renewMutex         *sync.Mutex
	signedCert         *auth.SignedCertificate
	tlsCert            tls.Certificate
	signedCertDuration time.Duration
	kind               string
	logger             logger.Logger
	maxConnectionAge   *time.Duration
	authToken          string
	apiSpec            config.APISpec
	proxy              messaging.Proxy
}

var (
	apiServerLogger      = logger.NewLogger("app.runtime.grpc.api")
	internalServerLogger = logger.NewLogger("app.runtime.grpc.internal")
)

// NewAPIServer returns a new user facing Bhojpur Application gRPC API server.
func NewAPIServer(api API, config ServerConfig, tracingSpec config.TracingSpec, metricSpec config.MetricSpec, apiSpec config.APISpec, proxy messaging.Proxy) Server {
	return &server{
		api:         api,
		config:      config,
		tracingSpec: tracingSpec,
		metricSpec:  metricSpec,
		kind:        apiServer,
		logger:      apiServerLogger,
		authToken:   auth.GetAPIToken(),
		apiSpec:     apiSpec,
		proxy:       proxy,
	}
}

// NewInternalServer returns a new gRPC server for Bhojpur Application runtime-to-runtime communications.
func NewInternalServer(api API, config ServerConfig, tracingSpec config.TracingSpec, metricSpec config.MetricSpec, authenticator auth.Authenticator, proxy messaging.Proxy) Server {
	return &server{
		api:              api,
		config:           config,
		tracingSpec:      tracingSpec,
		metricSpec:       metricSpec,
		authenticator:    authenticator,
		renewMutex:       &sync.Mutex{},
		kind:             internalServer,
		logger:           internalServerLogger,
		maxConnectionAge: getDefaultMaxAgeDuration(),
		proxy:            proxy,
	}
}

func getDefaultMaxAgeDuration() *time.Duration {
	d := time.Second * defaultMaxConnectionAgeSeconds
	return &d
}

// StartNonBlocking starts a new Bhojpur Application server in a goroutine.
func (s *server) StartNonBlocking() error {
	var listeners []net.Listener
	if s.config.UnixDomainSocket != "" && s.kind == apiServer {
		socket := fmt.Sprintf("%s/app-%s-grpc.socket", s.config.UnixDomainSocket, s.config.AppID)
		l, err := net.Listen("unix", socket)
		if err != nil {
			return err
		}
		listeners = append(listeners, l)
	} else {
		for _, apiListenAddress := range s.config.APIListenAddresses {
			l, err := net.Listen("tcp", fmt.Sprintf("%s:%v", apiListenAddress, s.config.Port))
			if err != nil {
				s.logger.Warnf("Failed to listen on %v:%v with error: %v", apiListenAddress, s.config.Port, err)
			} else {
				listeners = append(listeners, l)
			}
		}
	}

	if len(listeners) == 0 {
		return errors.Errorf("could not listen on any endpoint")
	}

	for _, listener := range listeners {
		// server is created in a loop because each instance
		// has a handle on the underlying listener.
		server, err := s.getGRPCServer()
		if err != nil {
			return err
		}
		s.servers = append(s.servers, server)

		if s.kind == internalServer {
			internalv1pb.RegisterServiceInvocationServer(server, s.api)
		} else if s.kind == apiServer {
			runtimev1pb.RegisterApplicationServer(server, s.api)
		}

		go func(server *grpc_go.Server, l net.Listener) {
			if err := server.Serve(l); err != nil {
				s.logger.Fatalf("gRPC serve error: %v", err)
			}
		}(server, listener)
	}
	return nil
}

func (s *server) Close() error {
	for _, server := range s.servers {
		// This calls `Close()` on the underlying listener.
		server.GracefulStop()
	}

	return nil
}

func (s *server) generateWorkloadCert() error {
	s.logger.Info("sending workload csr request to sentry")
	signedCert, err := s.authenticator.CreateSignedWorkloadCert(s.config.AppID, s.config.NameSpace, s.config.TrustDomain)
	if err != nil {
		return errors.Wrap(err, "error from authenticator CreateSignedWorkloadCert")
	}
	s.logger.Info("certificate signed successfully")

	tlsCert, err := tls.X509KeyPair(signedCert.WorkloadCert, signedCert.PrivateKeyPem)
	if err != nil {
		return errors.Wrap(err, "error creating x509 Key Pair")
	}

	s.signedCert = signedCert
	s.tlsCert = tlsCert
	s.signedCertDuration = signedCert.Expiry.Sub(time.Now().UTC())
	return nil
}

func (s *server) getMiddlewareOptions() []grpc_go.ServerOption {
	opts := []grpc_go.ServerOption{}
	intr := []grpc_go.UnaryServerInterceptor{}
	intrStream := []grpc_go.StreamServerInterceptor{}

	if len(s.apiSpec.Allowed) > 0 {
		s.logger.Info("enabled API access list on gRPC server")
		intr = append(intr, setAPIEndpointsMiddlewareUnary(s.apiSpec.Allowed))
	}

	if s.authToken != "" {
		s.logger.Info("enabled token authentication on gRPC server")
		intr = append(intr, setAPIAuthenticationMiddlewareUnary(s.authToken, auth.APITokenHeader))
	}

	if diag_utils.IsTracingEnabled(s.tracingSpec.SamplingRate) {
		s.logger.Info("enabled gRPC tracing middleware")
		intr = append(intr, diag.GRPCTraceUnaryServerInterceptor(s.config.AppID, s.tracingSpec))

		if s.proxy != nil {
			intrStream = append(intrStream, diag.GRPCTraceStreamServerInterceptor(s.config.AppID, s.tracingSpec))
		}
	}

	if s.metricSpec.Enabled {
		s.logger.Info("enabled gRPC metrics middleware")
		intr = append(intr, diag.DefaultGRPCMonitoring.UnaryServerInterceptor())
	}

	chain := grpc_middleware.ChainUnaryServer(
		intr...,
	)
	opts = append(
		opts,
		grpc_go.UnaryInterceptor(chain),
	)

	if s.proxy != nil {
		chainStream := grpc_middleware.ChainStreamServer(
			intrStream...,
		)

		opts = append(opts, grpc_go.StreamInterceptor(chainStream))
	}

	return opts
}

func (s *server) getGRPCServer() (*grpc_go.Server, error) {
	opts := s.getMiddlewareOptions()
	if s.maxConnectionAge != nil {
		opts = append(opts, grpc_go.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: *s.maxConnectionAge}))
	}

	if s.authenticator != nil {
		err := s.generateWorkloadCert()
		if err != nil {
			return nil, err
		}

		// nolint:gosec
		tlsConfig := tls.Config{
			ClientCAs:  s.signedCert.TrustChain,
			ClientAuth: tls.RequireAndVerifyClientCert,
			GetCertificate: func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
				return &s.tlsCert, nil
			},
		}
		ta := credentials.NewTLS(&tlsConfig)

		opts = append(opts, grpc_go.Creds(ta))
		go s.startWorkloadCertRotation()
	}

	opts = append(opts, grpc_go.MaxRecvMsgSize(s.config.MaxRequestBodySize*1024*1024), grpc_go.MaxSendMsgSize(s.config.MaxRequestBodySize*1024*1024), grpc_go.MaxHeaderListSize(uint32(s.config.ReadBufferSize*1024)))

	if s.proxy != nil {
		opts = append(opts, grpc_go.UnknownServiceHandler(s.proxy.Handler()))
	}

	return grpc_go.NewServer(opts...), nil
}

func (s *server) startWorkloadCertRotation() {
	s.logger.Infof("starting workload cert expiry watcher. current cert expires on: %s", s.signedCert.Expiry.String())

	ticker := time.NewTicker(certWatchInterval)

	for range ticker.C {
		s.renewMutex.Lock()
		renew := shouldRenewCert(s.signedCert.Expiry, s.signedCertDuration)
		if renew {
			s.logger.Info("renewing certificate: requesting new cert and restarting gRPC server")

			err := s.generateWorkloadCert()
			if err != nil {
				s.logger.Errorf("error starting server: %s", err)
			}
			diag.DefaultMonitoring.MTLSWorkLoadCertRotationCompleted()
		}
		s.renewMutex.Unlock()
	}
}

func shouldRenewCert(certExpiryDate time.Time, certDuration time.Duration) bool {
	expiresIn := certExpiryDate.Sub(time.Now().UTC())
	expiresInSeconds := expiresIn.Seconds()
	certDurationSeconds := certDuration.Seconds()

	percentagePassed := 100 - ((expiresInSeconds / certDurationSeconds) * 100)
	return percentagePassed >= renewWhenPercentagePassed
}
