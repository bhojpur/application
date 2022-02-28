package sentry

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

	"github.com/pkg/errors"

	"github.com/bhojpur/application/pkg/sentry/ca"
	"github.com/bhojpur/application/pkg/sentry/config"
	"github.com/bhojpur/application/pkg/sentry/identity"
	"github.com/bhojpur/application/pkg/sentry/identity/kubernetes"
	"github.com/bhojpur/application/pkg/sentry/identity/selfhosted"
	k8s "github.com/bhojpur/application/pkg/sentry/kubernetes"
	"github.com/bhojpur/application/pkg/sentry/monitoring"
	"github.com/bhojpur/application/pkg/sentry/server"
	"github.com/bhojpur/service/pkg/utils/logger"
)

var log = logger.NewLogger("app.sentry")

type CertificateAuthority interface {
	Run(context.Context, config.SentryConfig, chan bool)
	Restart(ctx context.Context, conf config.SentryConfig)
}

type sentry struct {
	server    server.CAServer
	reloading bool
}

// NewSentryCA returns a new Sentry Certificate Authority instance.
func NewSentryCA() CertificateAuthority {
	return &sentry{}
}

// Run loads the trust anchors and issuer certs, creates a new CA and runs the CA server.
func (s *sentry) Run(ctx context.Context, conf config.SentryConfig, readyCh chan bool) {
	// Create CA
	certAuth, err := ca.NewCertificateAuthority(conf)
	if err != nil {
		log.Fatalf("error getting certificate authority: %s", err)
	}
	log.Info("certificate authority loaded")

	// Load the trust bundle
	err = certAuth.LoadOrStoreTrustBundle()
	if err != nil {
		log.Fatalf("error loading trust root bundle: %s", err)
	}
	log.Infof("trust root bundle loaded. issuer cert expiry: %s", certAuth.GetCACertBundle().GetIssuerCertExpiry().String())
	monitoring.IssuerCertExpiry(certAuth.GetCACertBundle().GetIssuerCertExpiry())

	// Create identity validator
	v, err := createValidator()
	if err != nil {
		log.Fatalf("error creating validator: %s", err)
	}
	log.Info("validator created")

	// Run the CA server
	s.server = server.NewCAServer(certAuth, v)

	go func() {
		<-ctx.Done()
		log.Info("Bhojpur Application Sentry Certificate Authority is shutting down")
		s.server.Shutdown() // nolint: errcheck
	}()

	if readyCh != nil {
		readyCh <- true
		s.reloading = false
	}

	log.Infof("Bhojpur Application Sentry Certificate Authority is running, protecting ya'll")
	err = s.server.Run(conf.Port, certAuth.GetCACertBundle())
	if err != nil {
		log.Fatalf("error starting Bhojpur Application runtime gRPC server: %s", err)
	}
}

func createValidator() (identity.Validator, error) {
	if config.IsKubernetesHosted() {
		// we're in Kubernetes, create client and init a new serviceaccount token validator
		kubeClient, err := k8s.GetClient()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create kubernetes client")
		}
		return kubernetes.NewValidator(kubeClient), nil
	}
	return selfhosted.NewValidator(), nil
}

func (s *sentry) Restart(ctx context.Context, conf config.SentryConfig) {
	if s.reloading {
		return
	}
	s.reloading = true

	s.server.Shutdown()
	go s.Run(ctx, conf, nil)
}
