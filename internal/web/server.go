package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/groundsgg/b3/internal/web/middleware"
	"github.com/groundsgg/b3/internal/web/pages"
	"github.com/groundsgg/b3/internal/web/routers"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	pages      pages.Pages
}

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

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping webserver")
	err := s.httpServer.Shutdown(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		s.logger.Info("kill web server")
		return s.httpServer.Close()
	}

	return err
}

func NewServer(cfg ServerConfig) *Server {
	server := &Server{
		httpServer: &http.Server{
			Addr: cfg.ListenAddr,
		},
		logger: cfg.Logger,
		pages:  cfg.Pages,
	}

	mw := middleware.Combine(
		middleware.CORS(cfg.BaseURL),
		middleware.HTTPLogging(cfg.Logger),
		middleware.PagesContext(server.pages),
	)

	server.httpServer.Handler = mw(routers.CoreRouter())

	return server
}
