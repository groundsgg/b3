// SPDX-License-Identifier: AGPL-3.0-or-later
package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/groundsgg/b3/internal/web/request"
)

func verifyToken(tokenString string, sessionKey []byte) (request.SessionInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return sessionKey, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithJSONNumber())

	if err != nil {
		return request.SessionInfo{}, err
	}

	if !token.Valid {
		return request.SessionInfo{}, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return request.SessionInfo{}, fmt.Errorf("invalid token claims")
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return request.SessionInfo{}, fmt.Errorf("invalid username claim")
	}

	//pl, err := parsePL(claims["pl"])
	plVal, ok := claims["pl"]
	if !ok {
		return request.SessionInfo{}, fmt.Errorf("missing permission level claim")
	}

	plN, ok := plVal.(json.Number)
	if !ok {
		return request.SessionInfo{}, fmt.Errorf("invalid permission level claim: %T", plVal)
	}

	pl, err := plN.Int64()
	if err != nil {
		return request.SessionInfo{}, fmt.Errorf("invalid permission level claim: %w", err)
	}

	return request.SessionInfo{
		Username:        username,
		PermissionLevel: request.PermissionLevel(pl),
	}, nil
}

// Session validates JWT cookies and attaches session info to the request.
func Session(key []byte) Middleware {
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

			if req.Session.Username == "" {
				req.Session = request.SessionInfo{
					PermissionLevel: request.GROUP_GUEST,
					Sign: func(tokenID, username string, pl request.PermissionLevel) (string, error) {
						claims := jwt.MapClaims{
							"sub":      tokenID,
							"exp":      time.Now().Add(time.Hour * 24).Unix(),
							"iat":      time.Now().Unix(),
							"username": username,
							"pl":       pl,
						}

						token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
						signedToken, err := token.SignedString(key)
						if err != nil {
							return "", err
						}
						return signedToken, nil
					},
				}
			}

			req.Logger.Debug("user session info",
				"username", req.Session.Username,
				"pl", req.Session.PermissionLevel,
			)

			next(req)
		}
	}
}
