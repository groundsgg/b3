package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/groundsgg/b3/internal/web/request"
)

func verifyToken(tokenString, sessionKey string) (*request.SessionInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return sessionKey, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return &request.SessionInfo{
		PermissionGroup: request.VIEWER,
	}, nil
}

func Session(key string) Middleware {
	return func(next Handler) Handler {
		return func(req *request.Request) {
			if c, err := req.OriginalRequest.Cookie("auth_token"); err == nil {
				sesInfo, err := verifyToken(c.Value, key)
				if err != nil {
					req.Logger.Warn("token validation failed",
						"err", err,
					)
				}

				req.Session = sesInfo
			}

			req.Logger.Debug("user session info",
				"session", req.Session,
			)

			next(req)
		}
	}
}
