package web

import (
	"log/slog"
	"net/http"

	"github.com/groundsgg/b3/internal/auth"
	"github.com/groundsgg/b3/internal/web/pages"
)

type ServerConfig struct {
	ListenAddr  string
	Logger      *slog.Logger
	BaseURL     string
	SessionKey  []byte
	Pages       pages.Pages
	AuthHandler auth.AuthHandler
}

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	pages      pages.Pages
}
