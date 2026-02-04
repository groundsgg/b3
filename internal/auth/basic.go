package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/groundsgg/b3/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// Type returns BASIC_AUTH for the basic auth handler.
func (h *basicHandler) Type() AuthType {
	return BASIC_AUTH
}

// PreLogin is a no-op for basic auth.
func (h *basicHandler) PreLogin(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// VerifyToken validates the auth token and returns the associated user info.
func (h *basicHandler) VerifyToken(token string) (*UserInfo, error) {
	claims, err := h.jwtH.verify(token)
	if err != nil {
		return nil, err
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return nil, err
	}

	username, err := claims.GetUsername()
	if err != nil {
		return nil, err
	}

	pl, err := claims.GetPermissionLevel()
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		SessionID:       subject,
		Username:        username,
		PermissionLevel: pl,
	}, nil
}

// LoginCallback validates the submitted basic auth credentials.
func (h *basicHandler) LoginCallback(w http.ResponseWriter, r *http.Request) LoginCallbackResult {
	// Limit request body size to reduce DoS risk from large form posts.
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10) // 8 KiB
	if err := r.ParseForm(); err != nil {
		return LoginCallbackResult{
			ErrorMessage: "invalid request",
		}
	}
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if username != "" && password != "" {
		if user := h.users[strings.ToLower(username)]; user != nil {
			if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err == nil {
				tokenID := uuid.NewString()
				token, err := h.jwtH.sign(tokenID, NewSessionClaims(user.Name, user.PermissionLevel), time.Hour*24)
				if err != nil {
					return LoginCallbackResult{
						ErrorMessage: "failed to create a session",
					}
				}
				return LoginCallbackResult{
					Success:         true,
					Username:        user.Name,
					PermissionLevel: user.PermissionLevel,
					Token:           token,
					SessionID:       tokenID,
				}
			}
		}
	}

	return LoginCallbackResult{
		ErrorMessage: "invalid username or password",
	}
}

func getBasicAuthHandler() (AuthHandler, error) {

	users := make(map[string]*basicUser)

	usersList := config.GetConfig().Auth.Basic.Users
	for _, rawUser := range usersList {
		args := strings.Split(rawUser, ":")
		if len(args) != 3 {
			return nil, fmt.Errorf("invalid user format '%s', use name:permission-level:password-hash", rawUser)
		}

		permLvl, err := strconv.ParseUint(args[1], 10, 8)
		if err != nil {
			return nil, err
		}

		users[strings.ToLower(args[0])] = &basicUser{
			Name:            args[0],
			PermissionLevel: uint8(permLvl),
			PasswordHash:    []byte(args[2]),
		}
	}

	return &basicHandler{
		users: users,
		jwtH: &jwtTokenHandler{
			signKey: []byte(config.GetConfig().Web.SessionSignKey),
		},
	}, nil
}
