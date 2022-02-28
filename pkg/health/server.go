package health

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
	"fmt"
	"net/http"
	"time"

	"github.com/bhojpur/service/pkg/utils/logger"
)

// Server is the interface for the Bhojpur Application runtime healthz server.
type Server interface {
	Run(context.Context, int) error
	Ready()
	NotReady()
}

type server struct {
	ready bool
	log   logger.Logger
}

// NewServer returns a new Bhojpur Application runtime healthz server.
func NewServer(log logger.Logger) Server {
	return &server{
		log: log,
	}
}

// Ready sets a ready state for the endpoint handlers.
func (s *server) Ready() {
	s.ready = true
}

// NotReady sets a not ready state for the endpoint handlers.
func (s *server) NotReady() {
	s.ready = false
}

// Run starts a net/http server with a healthz endpoint.
func (s *server) Run(ctx context.Context, port int) error {
	router := http.NewServeMux()
	router.Handle("/healthz", s.healthz())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	doneCh := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			s.log.Info("Bhojpur Application Healthz server is shutting down")
			shutdownCtx, cancel := context.WithTimeout(
				context.Background(),
				time.Second*5,
			)
			defer cancel()
			srv.Shutdown(shutdownCtx) // nolint: errcheck
		case <-doneCh:
		}
	}()

	s.log.Infof("Bhojpur Application Healthz server is listening on %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		s.log.Errorf("Healthz server error: %s", err)
	}
	close(doneCh)
	return err
}

// healthz is a health endpoint handler.
func (s *server) healthz() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var status int
		if s.ready {
			status = http.StatusOK
		} else {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)
	})
}
