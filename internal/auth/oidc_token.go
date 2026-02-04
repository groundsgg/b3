package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

const oidcProactiveRefreshWindow = time.Minute

func (h *oidcHandler) parseAuthToken(rawToken string) (*JWTClaims, error) {
	rawToken, err := h.jwtHandler.decrypt(rawToken)
	if err != nil {
		return nil, err
	}

	claims, err := h.jwtHandler.verify(rawToken, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}

	if err := claims.Validate(); err != nil {
		return nil, err
	}

	return claims, nil
}

func (h *oidcHandler) buildAuthToken(user *OIDCUserInfo, refreshToken string, expiry time.Time) (string, error) {
	pl, err := h.groupToPL(user.B3Group)
	if err != nil {
		return "", err
	}

	if expiry.IsZero() {
		return "", fmt.Errorf("missing token expiry")
	}

	claims := NewSessionClaims(user.Name, pl).WithRefreshToken(refreshToken)

	rawToken, err := h.jwtHandler.sign(user.Subject, claims, time.Until(expiry))
	if err != nil {
		return "", err
	}
	encToken := h.jwtHandler.encrypt(rawToken)
	return encToken, nil
}

func (h *oidcHandler) groupToPL(group string) (uint8, error) {
	var pl uint8
	switch strings.ToLower(group) {
	case "viewer":
		pl = 5
	case "editor":
		pl = 10
	case "admin":
		pl = 20
	default:
		return 0, fmt.Errorf("unknown b3 group: %s", group)
	}
	return pl, nil
}

func (h *oidcHandler) updateToken(refreshToken string) (*UserInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	ts := h.oauthCfg.TokenSource(ctx, token)
	t, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("token refreshing error: %w", err)
	}

	rawIDToken, ok := t.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("token verification error: no id token")
	}

	idToken, err := h.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("token verification error: %w", err)
	}

	var user OIDCUserInfo
	if err := idToken.Claims(&user); err != nil {
		return nil, fmt.Errorf("userinfo decode failed: %w", err)
	}

	pl, err := h.groupToPL(user.B3Group)
	if err != nil {
		return nil, err
	}

	ui := &UserInfo{Username: user.Name, PermissionLevel: pl, SessionID: idToken.Subject}

	newRefreshToken := t.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = refreshToken
	}

	expiry := idToken.Expiry
	if expiry.IsZero() {
		expiry = t.Expiry
	}

	authToken, err := h.buildAuthToken(&user, newRefreshToken, expiry)
	if err != nil {
		return nil, err
	}
	ui.NewToken = authToken

	return ui, nil
}

// VerifyToken verifies the provided OIDC token, refreshing it if needed, and returns user info.
func (h *oidcHandler) VerifyToken(rawToken string) (*UserInfo, error) {
	claims, err := h.parseAuthToken(rawToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	if claims.ExpiresAt == nil {
		return nil, fmt.Errorf("invalid token: missing exp claim")
	}

	now := time.Now()
	if now.After(claims.ExpiresAt.Time.Add(jwtClockLeeway)) {
		if claims.RefreshToken == "" {
			return nil, fmt.Errorf("token expired and missing refresh token")
		}
		return h.updateToken(claims.RefreshToken)
	}

	if claims.RefreshToken != "" && claims.ExpiresAt.Time.Sub(now) < oidcProactiveRefreshWindow {
		return h.updateToken(claims.RefreshToken)
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
		SessionID:       claims.Subject,
		Username:        username,
		PermissionLevel: pl,
	}, nil
}
