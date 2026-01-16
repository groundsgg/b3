// SPDX-License-Identifier: AGPL-3.0-or-later
package web

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/groundsgg/b3/internal/web/middleware"
	"github.com/groundsgg/b3/internal/web/request"
	"github.com/groundsgg/b3/internal/web/routers"
	"github.com/groundsgg/b3/internal/web/routers/auth"
	"github.com/groundsgg/b3/internal/web/routers/core"
)

// Start runs the HTTP server and blocks until it stops.
func (s *Server) Start() error {
	s.logger.Info("starting web server",
		"addr", s.httpServer.Addr,
	)
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Stop gracefully shuts down the HTTP server with the provided context.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping web server")
	err := s.httpServer.Shutdown(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		s.logger.Info("force closing web server")
		return s.httpServer.Close()
	}

	return err
}

// NewServer builds a configured web server with routes and middleware.
func NewServer(cfg ServerConfig) (*Server, error) {
	server := &Server{
		httpServer: &http.Server{
			Addr:              cfg.ListenAddr,
			ReadHeaderTimeout: time.Second * 5,
			IdleTimeout:       time.Second * 60,
			MaxHeaderBytes:    1 << 20, // 1 MiB
		},
		logger: cfg.Logger,
		pages:  cfg.Pages,
	}

	mw := middleware.Combine(
		middleware.CORS(cfg.BaseURL),
		middleware.HTTPLogging(cfg.Logger),
		middleware.Session(cfg.SessionKey),
	)

	assetHandler, err := core.AssetsHandler()
	if err != nil {
		return nil, err
	}
	mainRouter := routers.NewRouter()
	mainRouter.Handle("GET /assets/", assetHandler)
	mainRouter.Handle("/auth/", auth.Handler())
	mainRouter.Handle("/", core.Home)
	mainRouter.Handle("/a", core.A)
	mainRouter.Handle("/b", core.B)

	server.httpServer.Handler = request.Parse(mw(mainRouter.Serve), server.pages, cfg.AuthHandler)

	return server, nil
}
