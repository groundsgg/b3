package web

import (
	"log/slog"

	"github.com/groundsgg/b3/internal/web/pages"
)

type ServerConfig struct {
	ListenAddr string
	Logger     *slog.Logger
	BaseURL    string
	Pages      pages.Pages
}
