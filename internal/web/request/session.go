// SPDX-License-Identifier: AGPL-3.0-or-later
package request

type PermissionLevel uint8

const (
	GROUP_GUEST  PermissionLevel = 0
	GROUP_VIEWER PermissionLevel = 5
	GROUP_EDITOR PermissionLevel = 10
	GROUP_ADMIN  PermissionLevel = 20
)

type SignRequest func(tokenID, username string, pl PermissionLevel) (string, error)

type SessionInfo struct {
	Username        string
	PermissionLevel PermissionLevel
	Sign            SignRequest
}

// IsAuthenticated reports whether the session has a non-empty username.
func (r *Request) IsAuthenticated() bool {
	return r.Session.Username != ""
}

// HasPermission checks if the session meets the required permission level.
func (r *Request) HasPermission(requiredLevel PermissionLevel) bool {
	return r.Session.PermissionLevel >= requiredLevel
}
