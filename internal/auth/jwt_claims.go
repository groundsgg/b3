package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	jwt.RegisteredClaims

	Username        string `json:"username,omitempty"`
	PermissionLevel *uint8 `json:"permission_level,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
}

var _ jwt.ClaimsValidator = (*JWTClaims)(nil)

func NewSessionClaims(username string, permissionLevel uint8) JWTClaims {
	pl := permissionLevel
	return JWTClaims{
		Username:        username,
		PermissionLevel: &pl,
	}
}

func (c JWTClaims) WithRefreshToken(refreshToken string) JWTClaims {
	c.RefreshToken = refreshToken
	return c
}

func (c JWTClaims) Validate() error {
	if c.Subject == "" {
		return errors.New("missing sub claim")
	}
	if c.ExpiresAt == nil {
		return errors.New("missing exp claim")
	}
	if c.IssuedAt == nil {
		return errors.New("missing iat claim")
	}

	if c.IssuedAt.Time.After(time.Now().Add(jwtClockLeeway)) {
		return errors.New("iat claim is in the future")
	}

	if c.Username == "" {
		return errors.New("missing username claim")
	}
	if c.PermissionLevel == nil {
		return errors.New("missing permission_level claim")
	}
	return nil
}

func (c JWTClaims) GetUsername() (string, error) {
	if c.Username == "" {
		return "", errors.New("missing username claim")
	}
	return c.Username, nil
}

func (c JWTClaims) GetPermissionLevel() (uint8, error) {
	if c.PermissionLevel == nil {
		return 0, errors.New("missing permission_level claim")
	}
	return *c.PermissionLevel, nil
}

func (c JWTClaims) GetRefreshToken() (string, error) {
	return c.RefreshToken, nil
}
