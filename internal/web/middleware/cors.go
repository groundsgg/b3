package middleware

import (
	"net/http"
	"net/url"

	"github.com/groundsgg/b3/internal/web/request"
)

// CORS sets CORS headers for the given origin and handles preflight requests.
func CORS(origin string) Middleware {
	if origin != "" && origin != "*" {
		if u, err := url.Parse(origin); err == nil && u.Scheme != "" && u.Host != "" {
			origin = u.Scheme + "://" + u.Host
		}
	}
	return func(next Handler) Handler {
		return func(req *request.Request) {
			r := req.OriginalRequest
			w := req.OriginalWriter
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			// Not necessary, but maybe for the future
			w.Header().Set("Vary", "Origin")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(req)
		}
	}
}
