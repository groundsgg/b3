// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/groundsgg/b3/pkg/gen"
	"golang.org/x/oauth2"
)

type oidcHandler struct {
	providerCfg     *oauth2.Config
	userEndpointURL string
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
	// CSRF State Cookie
	http.SetCookie(res, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
		Path:     "/",
	})

	url := h.providerCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
	)

	http.Redirect(res, req, url, http.StatusSeeOther)
	return nil
}

// LoginCallback handles the OIDC code exchange and user info lookup.
func (h *oidcHandler) LoginCallback(res http.ResponseWriter, req *http.Request) LoginCallbackResult {
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

	http.SetCookie(res, &http.Cookie{
		Name:     "oidc_state",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	code := req.URL.Query().Get("code")
	if code == "" {
		return LoginCallbackResult{ErrorMessage: "missing code"}
	}

	ctx, cancel := context.WithTimeout(rCTX, time.Second*10)
	defer cancel()

	token, err := h.providerCfg.Exchange(ctx, code)
	if err != nil {
		return LoginCallbackResult{ErrorMessage: "token exchange failed"}
	}

	client := h.providerCfg.Client(ctx, token)
	resp, err := client.Get(h.userEndpointURL)
	if err != nil {
		return LoginCallbackResult{ErrorMessage: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return LoginCallbackResult{ErrorMessage: fmt.Sprintf("invalid status code: %d", resp.StatusCode)}
	}

	var user OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return LoginCallbackResult{ErrorMessage: err.Error()}
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

	var d OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func getOIDCHandler(baseURL string) (AuthHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	discovery, err := fetchDiscovery(ctx,
		os.Getenv("AUTH_OIDC_DISCOVERY_URL"),
	)

	if err != nil {
		return nil, err
	}

	cfg := &oauth2.Config{
		ClientID:     os.Getenv("AUTH_OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH_OIDC_CLIENT_SECRET"),
		RedirectURL:  baseURL + "/auth/code",
		Endpoint: oauth2.Endpoint{
			AuthURL:  discovery.AuthorizationEndpoint,
			TokenURL: discovery.TokenEndpoint,
		},
		Scopes: []string{"openid", "profile", "b3"},
	}

	return &oidcHandler{
		providerCfg:     cfg,
		userEndpointURL: discovery.UserinfoEndpoint,
	}, nil
}
