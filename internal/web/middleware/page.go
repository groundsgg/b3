package middleware

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/pages"
)

func PagesContext(p pages.Pages) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := pages.WithPages(r.Context(), p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
