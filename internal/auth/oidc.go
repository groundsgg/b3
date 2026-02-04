package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/groundsgg/b3/internal/config"
	"github.com/groundsgg/b3/internal/web/cookie"
	"github.com/groundsgg/b3/pkg/gen"
	"github.com/groundsgg/b3/pkg/log"
	"golang.org/x/oauth2"
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

	// CSRF State cookie
	cookie.Set(res, cookie.OIDC_STATE, state, 5*time.Minute)

	// PKCE code_verifier cookie
	cookie.Set(res, cookie.OIDC_VERIFIER, verifier, 5*time.Minute)

	// Nonce cookie
	cookie.Set(res, cookie.OIDC_NONCE, nonce, 5*time.Minute)

	url := h.oauthCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", pkceChallengeS256(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(res, req, url, http.StatusSeeOther)
	return nil
}

func (h *oidcHandler) getCookies(res http.ResponseWriter, req *http.Request) (oidcCookies, error) {
	cookies := oidcCookies{}
	queryState := req.URL.Query().Get("state")
	if queryState == "" {
		return cookies, errors.New("missing state")
	}

	cookie.Delete(res, cookie.OIDC_STATE)
	cookie.Delete(res, cookie.OIDC_VERIFIER)
	cookie.Delete(res, cookie.OIDC_NONCE)

	cState, err := req.Cookie(cookie.OIDC_STATE)
	if err != nil {
		return cookies, errors.New("missing state cookie")
	}
	cookies.State = cState.Value

	cVerifier, err := req.Cookie(cookie.OIDC_VERIFIER)
	if err != nil {
		return cookies, errors.New("missing pkce verifier")
	}
	cookies.Verifier = cVerifier.Value

	cNonce, err := req.Cookie(cookie.OIDC_NONCE)
	if err != nil {
		return cookies, errors.New("missing nonce cookie")
	}
	cookies.Nonce = cNonce.Value

	if cState.Value != queryState {
		return cookies, errors.New("invalid state")
	}

	return cookies, nil
}

func (h *oidcHandler) exchangeUserInfo(ctx context.Context, code, codeVerifier, nonce string) (*OIDCUserInfo, *oauth2.Token, time.Time, error) {
	rctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	oauthToken, err := h.oauthCfg.Exchange(rctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("token exchange failed: %w", err)
	}

	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		return nil, nil, time.Time{}, errors.New("no id_token field in oauth2 token")
	}

	idToken, err := h.verifier.Verify(rctx, rawIDToken)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("failed to verify ID Token: %w", err)
	}

	if idToken.Nonce != nonce {
		return nil, nil, time.Time{}, errors.New("nonce did not match")
	}

	var user OIDCUserInfo
	if err := idToken.Claims(&user); err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("userinfo decode failed: %w", err)
	}

	expiry := idToken.Expiry
	if expiry.IsZero() {
		expiry = oauthToken.Expiry
	}

	return &user, oauthToken, expiry, nil
}

// LoginCallback handles the OIDC code exchange and user info lookup.
func (h *oidcHandler) LoginCallback(res http.ResponseWriter, req *http.Request) LoginCallbackResult {
	logger := log.LoggerFromContext(req.Context())

	cookies, err := h.getCookies(res, req)
	if err != nil {
		return LoginCallbackResult{ErrorMessage: err.Error()}
	}

	code := req.URL.Query().Get("code")
	if code == "" {
		return LoginCallbackResult{ErrorMessage: "missing code"}
	}

	user, oauthToken, expiry, err := h.exchangeUserInfo(req.Context(), code, cookies.Verifier, cookies.Nonce)
	if err != nil {
		logger.Warn("oidc verification failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "verification failed"}
	}

	authToken, err := h.buildAuthToken(user, oauthToken.RefreshToken, expiry)
	if err != nil {
		logger.Warn("oidc token parsing failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "internal error"}
	}

	pl, err := h.groupToPL(user.B3Group)
	if err != nil {
		logger.Warn("oidc group mapping failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "internal error"}
	}

	return LoginCallbackResult{Success: true,
		Token:           authToken,
		Username:        user.Name,
		PermissionLevel: pl,
		SessionID:       user.Subject,
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
		Scopes: []string{
			oidc.ScopeOpenID,
			"profile",
			"offline_access",
			"b3",
		},
	}
	jwtH, err := newJWTHandler([]byte(config.GetConfig().Web.SessionSignKey),
		[]byte(config.GetConfig().Web.SessionEncryptKey))

	if err != nil {
		return nil, err
	}

	return &oidcHandler{
		oauthCfg:   oauthCfg,
		verifier:   verifier,
		provider:   provider,
		jwtHandler: jwtH,
	}, nil
}
