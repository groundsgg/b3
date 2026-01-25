package auth

import (
	"net/http"

	"github.com/groundsgg/b3/internal/auth"
	"github.com/groundsgg/b3/internal/config"
	"github.com/groundsgg/b3/internal/web/request"
	"github.com/groundsgg/b3/internal/web/routers"
	"github.com/groundsgg/b3/pkg/log"
)

const (
	AUTH_TOKEN string = "auth_token"
)

func getLogin(req *request.Request) {
	ctx := log.WithLogger(req.OriginalRequest.Context(), req.Logger)
	err := req.AuthHandler.PreLogin(req.OriginalWriter, req.OriginalRequest.WithContext(ctx))

	if err != nil {
		req.Logger.Warn("pre-login failed", "err", err)
		err := req.PrintError(request.ErrorData{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		if err != nil {
			req.Logger.Error("template rendering error", "err", err)
		}
		return
	}

	switch req.AuthHandler.Type() {
	case auth.BASIC_AUTH:
		req.OriginalWriter.WriteHeader(http.StatusOK)
		err := req.Print("pages.login", "Login", request.ErrorData{})
		if err != nil {
			req.Logger.Error("template rendering error", "err", err)
		}
	case auth.OIDC:
	default:
		err := req.PrintError(request.ErrorData{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		if err != nil {
			req.Logger.Error("template rendering error", "err", err)
		}
	}
}

func postLogin(req *request.Request) {
	ctx := log.WithLogger(req.OriginalRequest.Context(), req.Logger)
	result := req.AuthHandler.LoginCallback(req.OriginalWriter, req.OriginalRequest.WithContext(ctx))

	if result.Success {
		req.Logger.Info("created user session",
			"username", result.Username,
			"pl", result.PermissionLevel,
		)

		secure := config.IsSecureConnection()
		sameSite := http.SameSiteLaxMode
		if secure {
			sameSite = http.SameSiteNoneMode
		}

		// set token
		http.SetCookie(req.OriginalWriter, &http.Cookie{
			Name:     AUTH_TOKEN,
			Value:    result.Token,
			HttpOnly: true,
			Path:     "/",
			SameSite: sameSite,
			Secure:   secure,
			MaxAge:   60 * 60 * 24, // 1 day
		})

		http.Redirect(req.OriginalWriter, req.OriginalRequest, "/", http.StatusSeeOther)
		return
	}

	if req.AuthHandler.Type() == auth.BASIC_AUTH {
		req.OriginalWriter.WriteHeader(http.StatusOK)
		err := req.Print("pages.login", "Login", request.ErrorData{
			Message: result.ErrorMessage,
		})
		if err != nil {
			req.Logger.Error("template rendering error", "err", err)
		}
	} else if req.AuthHandler.Type() == auth.OIDC {
		err := req.PrintError(request.ErrorData{
			Message: result.ErrorMessage,
			Code:    http.StatusBadRequest,
		})
		if err != nil {
			req.Logger.Error("template rendering error", "err", err)
		}
	} else {
		err := req.PrintError(request.ErrorData{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		})
		if err != nil {
			req.Logger.Error("template rendering error", "err", err)
		}
	}
}

func logout(req *request.Request) {
	secure := config.IsSecureConnection()
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(req.OriginalWriter, &http.Cookie{
		Name:     AUTH_TOKEN,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
		SameSite: sameSite,
		Secure:   secure,
	})
	http.Redirect(req.OriginalWriter, req.OriginalRequest, "/", http.StatusSeeOther)
}

func notFound(req *request.Request) {
	err := req.PrintNotFound()
	if err != nil {
		req.Logger.Error("template rendering error", "err", err)
	}
}

// Handler builds the auth router and returns its Serve handler.
func Handler() func(*request.Request) {
	authRouter := routers.NewRouter()

	authRouter.Handle("GET /auth/login", notAuthenticated(getLogin))
	authRouter.Handle("POST /auth/login", notAuthenticated(postLogin))
	authRouter.Handle("GET /auth/code", notAuthenticated(postLogin))
	authRouter.Handle("GET /auth/logout", authenticated(logout))
	authRouter.Handle("/auth/", notFound)

	return authRouter.Serve
}
