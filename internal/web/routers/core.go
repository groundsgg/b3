package routers

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/pages"
	"github.com/groundsgg/b3/pkg/log"
)

func coreHome(w http.ResponseWriter, r *http.Request) {
	p := pages.PagesFromContext(r.Context())
	logger := log.LoggerFromContext(r.Context())

	w.WriteHeader(http.StatusOK)
	err := p.Render(w, "pages.home", pages.PageData{Title: "Welcome"})
	if err != nil {
		logger.Error("failed to parse template",
			"template", "pages.home",
			"err", err,
		)
	}
}

func a(w http.ResponseWriter, r *http.Request) {
	p := pages.PagesFromContext(r.Context())
	logger := log.LoggerFromContext(r.Context())

	w.WriteHeader(http.StatusOK)
	err := p.Render(w, "pages.a", pages.PageData{Title: "AAA"})
	if err != nil {
		logger.Error("failed to parse template",
			"template", "pages.a",
			"err", err,
		)
	}
}

func b(w http.ResponseWriter, r *http.Request) {
	p := pages.PagesFromContext(r.Context())
	logger := log.LoggerFromContext(r.Context())

	w.WriteHeader(http.StatusOK)
	err := p.Render(w, "pages.b", pages.PageData{Title: "BBB"})
	if err != nil {
		logger.Error("failed to parse template",
			"template", "pages.b",
			"err", err,
		)
	}
}

func CoreRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", coreHome)
	mux.HandleFunc("/a", a)
	mux.HandleFunc("/b", b)

	return mux
}
