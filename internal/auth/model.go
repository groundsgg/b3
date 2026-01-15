// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import "net/http"

type AuthType string

const (
	BASIC_AUTH AuthType = "basic_auth"
)

type LoginCallbackResult struct {
	Success         bool
	ErrorMessage    string
	Username        string
	PermissionLevel uint8
}

type AuthHandler interface {
	Type() AuthType
	PreLogin(r *http.Request)
	LoginCallback(r *http.Request) LoginCallbackResult
}
