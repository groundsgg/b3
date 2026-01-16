// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/middleware"
	"github.com/groundsgg/b3/internal/web/request"
)

func notAuthenticated(next middleware.Handler) middleware.Handler {
	return func(req *request.Request) {
		if req.IsAuthenticated() {
			http.Redirect(req.OriginalWriter, req.OriginalRequest, "/", http.StatusSeeOther)
			return
		}

		next(req)
	}
}

func authenticated(next middleware.Handler) middleware.Handler {
	return func(req *request.Request) {
		if !req.IsAuthenticated() {
			err := req.PrintError(request.ErrorData{
				Code:    http.StatusUnauthorized,
				Message: "unauthorized",
			})
			if err != nil {
				req.Logger.Error("template rendering error", "err", err)
			}
			return
		}

		next(req)
	}
}
