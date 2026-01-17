// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/groundsgg/b3/pkg/gen"
	"github.com/groundsgg/b3/pkg/log"
	"golang.org/x/oauth2"
)

type oidcHandler struct {
	providerCfg     *oauth2.Config
	userEndpointURL string
	secureCookie    bool
}

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
	verifier, err := gen.Hex(32)
	if err != nil {
		return err
	}

	sameSite := http.SameSiteLaxMode
	if h.secureCookie {
		sameSite = http.SameSiteNoneMode
	}

	// CSRF State Cookie
	http.SetCookie(res, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   h.secureCookie,
		MaxAge:   300, // 5 minutes
		Path:     "/",
	})
	// PKCE code_verifier cookie
	http.SetCookie(res, &http.Cookie{
		Name:     "oidc_verifier",
		Value:    verifier,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   h.secureCookie,
		MaxAge:   300, // 5 minutes
		Path:     "/",
	})

	url := h.providerCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", pkceChallengeS256(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(res, req, url, http.StatusSeeOther)
	return nil
}

// LoginCallback handles the OIDC code exchange and user info lookup.
func (h *oidcHandler) LoginCallback(res http.ResponseWriter, req *http.Request) LoginCallbackResult {
	logger := log.LoggerFromContext(req.Context())
	rCTX := req.Context()

	queryState := req.URL.Query().Get("state")
	if queryState == "" {
		return LoginCallbackResult{ErrorMessage: "missing state"}
	}

	cookie, err := req.Cookie("oidc_state")
	if err != nil {
		return LoginCallbackResult{ErrorMessage: "missing state cookie"}
	}

	if cookie.Value != queryState {
		return LoginCallbackResult{ErrorMessage: "invalid state"}
	}

	sameSite := http.SameSiteLaxMode
	if h.secureCookie {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(res, &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   h.secureCookie,
		Path:     "/",
	})

	verifierCookie, err := req.Cookie("oidc_verifier")
	if err != nil {
		return LoginCallbackResult{ErrorMessage: "missing pkce verifier"}
	}
	http.SetCookie(res, &http.Cookie{
		Name:     "oidc_verifier",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   h.secureCookie,
		Path:     "/",
	})

	code := req.URL.Query().Get("code")
	if code == "" {
		return LoginCallbackResult{ErrorMessage: "missing code"}
	}

	ctx, cancel := context.WithTimeout(rCTX, time.Second*10)
	defer cancel()

	token, err := h.providerCfg.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifierCookie.Value))
	if err != nil {
		logger.Warn("oidc token exchange failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "token exchange failed"}
	}

	client := h.providerCfg.Client(ctx, token)
	resp, err := client.Get(h.userEndpointURL)
	if err != nil {
		logger.Warn("oidc userinfo request failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "userinfo request failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("oidc userinfo request returned non-200", "status_code", resp.StatusCode)
		return LoginCallbackResult{ErrorMessage: "userinfo request failed"}
	}

	var user OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		logger.Warn("oidc userinfo decode failed", "err", err)
		return LoginCallbackResult{ErrorMessage: "userinfo decode failed"}
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

	return LoginCallbackResult{Success: true, Username: user.Name, PermissionLevel: pl}
}

func fetchDiscovery(ctx context.Context, wellKnownURL string) (*OIDCDiscovery, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", wellKnownURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery request failed: %d", resp.StatusCode)
	}

	var d OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func getOIDCHandler(baseURL string) (AuthHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	discoveryURL := os.Getenv("AUTH_OIDC_DISCOVERY_URL")
	if discoveryURL == "" {
		return nil, errors.New("please define AUTH_OIDC_DISCOVERY_URL")
	}

	discovery, err := fetchDiscovery(ctx, discoveryURL)

	if err != nil {
		return nil, err
	}

	cID := os.Getenv("AUTH_OIDC_CLIENT_ID")
	if cID == "" {
		return nil, errors.New("please define AUTH_OIDC_CLIENT_ID")
	}

	cSecret := os.Getenv("AUTH_OIDC_CLIENT_SECRET")
	if cSecret == "" {
		return nil, errors.New("please define AUTH_OIDC_CLIENT_SECRET")
	}

	cfg := &oauth2.Config{
		ClientID:     cID,
		ClientSecret: cSecret,
		RedirectURL:  strings.TrimSuffix(baseURL, "/") + "/auth/code",
		Endpoint: oauth2.Endpoint{
			AuthURL:  discovery.AuthorizationEndpoint,
			TokenURL: discovery.TokenEndpoint,
		},
		Scopes: []string{"openid", "profile", "b3"},
	}

	return &oidcHandler{
		providerCfg:     cfg,
		userEndpointURL: discovery.UserinfoEndpoint,
		secureCookie:    strings.HasPrefix(baseURL, "https"),
	}, nil
}
