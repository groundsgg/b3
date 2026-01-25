package middleware

import (
	"github.com/groundsgg/b3/internal/auth"
	"github.com/groundsgg/b3/internal/web/cookie"
	"github.com/groundsgg/b3/internal/web/request"
)

// Session validates JWT cookies and attaches session info to the request.
func Session(authHandler auth.AuthHandler) Middleware {
	return func(next Handler) Handler {
		return func(req *request.Request) {
			if c, err := req.OriginalRequest.Cookie(cookie.AUTH_TOKEN); err == nil {
				userInfo, err := authHandler.VerifyToken(c.Value)
				if err != nil {
					req.Logger.Warn("token validation failed",
						"err", err,
					)
					cookie.Delete(req.OriginalWriter, cookie.AUTH_TOKEN)
				} else {
					req.Session = request.SessionInfo{
						Username:        userInfo.Username,
						PermissionLevel: request.PermissionLevel(userInfo.PermissionLevel),
					}
					req.Logger = req.Logger.With("session_id", userInfo.SessionID)

					if userInfo.NewToken != "" {
						req.Logger.Info("user session refreshed",
							"username", userInfo.Username,
							"permission_level", userInfo.PermissionLevel,
						)
						cookie.Set(req.OriginalWriter, cookie.AUTH_TOKEN, userInfo.NewToken, cookie.AUTH_TOKEN_LIFETIME)
					}
				}
			}

			if req.Session.Username == "" {
				req.Session = request.SessionInfo{
					PermissionLevel: request.GROUP_GUEST,
				}
			}

			req.Logger.Debug("user session info",
				"username", req.Session.Username,
				"permission_level", req.Session.PermissionLevel,
			)

			next(req)
		}
	}
}
