// SPDX-License-Identifier: AGPL-3.0-or-later
package middleware

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/request"
)

// CORS sets CORS headers for the given origin and handles preflight requests.
func CORS(origin string) Middleware {
	return func(next Handler) Handler {
		return func(req *request.Request) {
			r := req.OriginalRequest
			w := req.OriginalWriter
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(req)
		}
	}
}
