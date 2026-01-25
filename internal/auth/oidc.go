package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/groundsgg/b3/internal/config"
	"github.com/groundsgg/b3/pkg/gen"
	"github.com/groundsgg/b3/pkg/log"
	"golang.org/x/oauth2"
)

const (
	OIDC_STATE    string = "oidc_state"
	OIDC_VERIFIER string = "oidc_verifier"
	OIDC_NONCE    string = "oidc_nonce"
)

func pkceChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// Type returns OIDC for the OIDC auth handler.
func (h *oidcHandler) Type() AuthType {
	return OIDC
}

// PreLogin starts the OIDC auth code flow and sets a state cookie.
func (h *oidcHandler) PreLogin(res http.ResponseWriter, req *http.Request) error {
	state, err := gen.Hex(32)
	if err != nil {
		return err
	}
	nonce, err := gen.Hex(32)
	if err != nil {
		return err
	}
	verifier, err := gen.Hex(32)
	if err != nil {
		return err
	}

	// CSRF State Cookie
	setCookie(res, OIDC_STATE, state, 5*time.Minute)

	// PKCE code_verifier cookie
	setCookie(res, OIDC_VERIFIER, verifier, 5*time.Minute)

	// PKCE code_verifier cookie
	setCookie(res, OIDC_NONCE, nonce, 5*time.Minute)

	url := h.oauthCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", pkceChallengeS256(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(res, req, url, http.StatusSeeOther)
	return nil
}

func (h *oidcHandler) VerifyToken(b64Token string) (*UserInfo, error) {
	rawToken, err := base64.RawStdEncoding.DecodeString(b64Token)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	var token oauth2.Token
	err = json.Unmarshal(rawToken, &token)
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}
	ts := h.oauthCfg.TokenSource(context.Background(), &token)
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

	var pl int
	switch strings.ToLower(user.B3Group) {
	case "viewer":
		pl = 5
	case "editor":
		pl = 10
	case "admin":
		pl = 20
	default:
		return nil, fmt.Errorf("unknown b3 group: %s", user.B3Group)
	}

	ui := &UserInfo{Username: user.Name, PermissionLevel: pl}
	if token.AccessToken != t.AccessToken {
		authToken, err := json.Marshal(t)
		if err == nil {
			ui.NewToken = base64.RawStdEncoding.EncodeToString(authToken)
		}
	}

	return ui, nil
}

func (h *oidcHandler) verifyCookies(res http.ResponseWriter, req *http.Request) (string, string, error) {
	queryState := req.URL.Query().Get("state")
	if queryState == "" {
		return "", "", errors.New("missing state")
	}

	setCookie(res, OIDC_STATE, "", -1)
	setCookie(res, OIDC_VERIFIER, "", -1)
	setCookie(res, OIDC_NONCE, "", -1)

	cookie, err := req.Cookie(OIDC_STATE)
	if err != nil {
		return "", "", errors.New("missing state cookie")
	}

	if cookie.Value != queryState {
		return "", "", errors.New("invalid state")
	}

	verifierCookie, err := req.Cookie(OIDC_VERIFIER)
	if err != nil {
		return "", "", errors.New("missing pkce verifier")
	}

	nonceCookie, err := req.Cookie(OIDC_NONCE)
	if err != nil {
		return "", "", errors.New("missing nonce cookie")
	}

	return verifierCookie.Value, nonceCookie.Value, nil
}

func (h *oidcHandler) exchangeUserinfo(ctx context.Context, code, verifierState, nonce string) (*OIDCUserInfo, *oauth2.Token, error) {
	rctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	oauthToken, err := h.oauthCfg.Exchange(rctx, code, oauth2.SetAuthURLParam("code_verifier", verifierState))
	if err != nil {
		return nil, nil, fmt.Errorf("token exchange failed: %w", err)
	}

	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		return nil, nil, errors.New("no id_token field in oauth2 token")
	}

	idToken, err := h.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify ID Token: %w", err)
	}

	if idToken.Nonce != nonce {
		return nil, nil, errors.New("nonce did not match")
	}

	userInfo, err := h.provider.UserInfo(ctx, oauth2.StaticTokenSource(oauthToken))
	if err != nil {
		return nil, nil, fmt.Errorf("userinfo request failed: %w", err)
	}

	var user OIDCUserInfo
	if err := userInfo.Claims(&user); err != nil {
		return nil, nil, fmt.Errorf("userinfo decode failed: %w", err)
	}

	return &user, oauthToken, nil
}

// LoginCallback handles the OIDC code exchange and user info lookup.
func (h *oidcHandler) LoginCallback(res http.ResponseWriter, req *http.Request) LoginCallbackResult {
	logger := log.LoggerFromContext(req.Context())

	verifierState, nonce, err := h.verifyCookies(res, req)
	if err != nil {
		return LoginCallbackResult{ErrorMessage: err.Error()}
	}

	code := req.URL.Query().Get("code")
	if code == "" {
		return LoginCallbackResult{ErrorMessage: "missing code"}
	}

	user, token, err := h.exchangeUserinfo(req.Context(), code, verifierState, nonce)
	if err != nil {
		logger.Warn("oidc verification failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "verification failed"}
	}

	var pl uint8
	switch strings.ToLower(user.B3Group) {
	case "viewer":
		pl = 5
	case "editor":
		pl = 10
	case "admin":
		pl = 20
	default:
		return LoginCallbackResult{ErrorMessage: "You are not authorized to log in"}
	}

	authToken, err := json.Marshal(token)
	if err != nil {
		logger.Warn("oidc token parsing failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "internal error"}
	}

	return LoginCallbackResult{Success: true,
		Token:           base64.RawStdEncoding.EncodeToString(authToken),
		Username:        user.Name,
		PermissionLevel: pl,
	}
}

func getOIDCHandler() (AuthHandler, error) {
	cfg := config.GetConfig().Auth.OIDC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}
	oidcCfg := &oidc.Config{
		ClientID: cfg.ClientID,
	}
	verifier := provider.Verifier(oidcCfg)

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  strings.TrimSuffix(config.GetConfig().Web.BaseURL, "/") + "/auth/code",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "offline_access", "b3"},
	}

	return &oidcHandler{
		oauthCfg: oauthCfg,
		verifier: verifier,
		provider: provider,
	}, nil
}
