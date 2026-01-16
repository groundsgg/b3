// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type basicUser struct {
	Name            string
	PasswordHash    []byte
	PermissionLevel uint8
}

type basicHandler struct {
	users map[string]*basicUser
}

// Type returns BASIC_AUTH for the basic auth handler.
func (h *basicHandler) Type() AuthType {
	return BASIC_AUTH
}

// PreLogin is a no-op for basic auth.
func (h *basicHandler) PreLogin(_ http.ResponseWriter, _ *http.Request) error {
	return nil
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
				return LoginCallbackResult{
					Success:         true,
					Username:        user.Name,
					PermissionLevel: user.PermissionLevel,
				}
			}
		}
	}

	return LoginCallbackResult{
		ErrorMessage: "invalid username or password",
	}
}

func getBasicAuthHandler() (AuthHandler, error) {
	rawUsers := os.Getenv("AUTH_BASIC_USERS")
	if rawUsers == "" {
		return nil, errors.New("please define users in AUTH_BASIC_USERS")
	}

	users := make(map[string]*basicUser)

	usersList := strings.Split(rawUsers, ",")
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
	}, nil
}
