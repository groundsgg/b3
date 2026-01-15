// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/groundsgg/b3/internal/auth"
	"github.com/groundsgg/b3/internal/web/request"
	"github.com/groundsgg/b3/internal/web/routers"
)

func getLogin(req *request.Request) {
	err := req.AuthHandler.PreLogin(req.OriginalWriter, req.OriginalRequest)

	if err != nil {
		req.PrintError(request.ErrorData{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	switch req.AuthHandler.Type() {
	case auth.BASIC_AUTH:
		req.OriginalWriter.WriteHeader(http.StatusOK)
		req.PrintOnly("pages.login", "Login", request.ErrorData{})
	case auth.OIDC:
	default:
		req.PrintError(request.ErrorData{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
	}
}

func postLogin(req *request.Request) {
	result := req.AuthHandler.LoginCallback(req.OriginalWriter, req.OriginalRequest)

	if result.Success {
		if req.Session.Sign != nil {
			tokenID := uuid.NewString()
			token, err := req.Session.Sign(tokenID, result.Username, request.PermissionLevel(result.PermissionLevel))
			if err != nil {
				req.Logger.Error("failed to create token", "err", err)
				req.PrintError(request.ErrorData{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				})
				return
			}

			req.Logger.Info("created user session",
				"username", result.Username,
				"pl", result.PermissionLevel,
				"token_id", tokenID,
			)

			// set token
			http.SetCookie(req.OriginalWriter, &http.Cookie{
				Name:     "auth_token",
				Value:    token,
				HttpOnly: true,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
				MaxAge:   60 * 60 * 24, // 1 day
			})
		}

		http.Redirect(req.OriginalWriter, req.OriginalRequest, "/", http.StatusSeeOther)
		return
	}

	if req.AuthHandler.Type() == auth.BASIC_AUTH {
		req.OriginalWriter.WriteHeader(http.StatusOK)
		req.PrintOnly("pages.login", "Login", request.ErrorData{
			Message: result.ErrorMessage,
		})
	} else if req.AuthHandler.Type() == auth.OIDC {
		req.PrintError(request.ErrorData{
			Message: result.ErrorMessage,
			Code:    http.StatusBadRequest,
		})
	} else {
		req.PrintError(request.ErrorData{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
	}
}

func logout(req *request.Request) {
	http.SetCookie(req.OriginalWriter, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})
	http.Redirect(req.OriginalWriter, req.OriginalRequest, "/", http.StatusSeeOther)
}

func notFound(req *request.Request) {
	req.PrintNotFound()
}

func Handler() func(*request.Request) {
	authRouter := routers.NewRouter()

	authRouter.Handle("GET /auth/login", notAuthenticated(getLogin))
	authRouter.Handle("POST /auth/login", notAuthenticated(postLogin))
	authRouter.Handle("GET /auth/code", notAuthenticated(postLogin))
	authRouter.Handle("GET /auth/logout", authenticated(logout))
	authRouter.Handle("/auth/", notFound)

	return authRouter.Serve
}
