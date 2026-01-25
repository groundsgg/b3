package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
)

func (h *oidcHandler) parseAuthToken(rawToken string) (*oauth2.Token, error) {
	jToken, err := base64.RawStdEncoding.DecodeString(rawToken)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	err = json.Unmarshal(jToken, &token)
	return &token, err
}

func (h *oidcHandler) buildAuthToken(token *oauth2.Token) (string, error) {
	jToken, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(jToken), nil
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

func (h *oidcHandler) VerifyToken(b64Token string) (*UserInfo, error) {
	token, err := h.parseAuthToken(b64Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	ts := h.oauthCfg.TokenSource(context.Background(), token)
	t, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("token refreshing error: %w", err)
	}

	idToken, err := h.verifier.Verify(context.Background(), t.AccessToken)
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

	ui := &UserInfo{Username: user.Name, PermissionLevel: pl}
	if token.AccessToken != t.AccessToken {
		authToken, err := h.buildAuthToken(t)
		if err == nil {
			ui.NewToken = authToken
		}
	}

	return ui, nil
}
