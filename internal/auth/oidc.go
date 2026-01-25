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

	"github.com/groundsgg/b3/internal/config"
	"github.com/groundsgg/b3/pkg/gen"
	"github.com/groundsgg/b3/pkg/log"
	"golang.org/x/oauth2"
)

const (
	OIDC_STATE    string = "oidc_state"
	OIDC_VERIFIER string = "oidc_verifier"
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
	verifier, err := gen.Hex(32)
	if err != nil {
		return err
	}

	// CSRF State Cookie
	setAuthCookie(res, OIDC_STATE, state, 5*time.Minute)

	// PKCE code_verifier cookie
	setAuthCookie(res, OIDC_VERIFIER, verifier, 5*time.Minute)

	url := h.providerCfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", pkceChallengeS256(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	http.Redirect(res, req, url, http.StatusSeeOther)
	return nil
}

func (h *oidcHandler) VerifyToken(token string) (*UserInfo, error) {
	return h.jwtH.verify(token)
}

func (h *oidcHandler) verifyCookies(res http.ResponseWriter, req *http.Request) (error, string) {
	queryState := req.URL.Query().Get("state")
	if queryState == "" {
		return errors.New("missing state"), ""
	}

	setAuthCookie(res, OIDC_STATE, "", -1)
	setAuthCookie(res, OIDC_VERIFIER, "", -1)

	cookie, err := req.Cookie(OIDC_STATE)
	if err != nil {
		return errors.New("missing state cookie"), ""
	}

	if cookie.Value != queryState {
		return errors.New("invalid state"), ""
	}

	verifierCookie, err := req.Cookie(OIDC_VERIFIER)
	if err != nil {
		return errors.New("missing pkce verifier"), ""
	}

	return nil, verifierCookie.Value
}

func (h *oidcHandler) exchangeUserinfo(ctx context.Context, code, verifierState string) (error, *OIDCUserInfo) {
	rctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	token, err := h.providerCfg.Exchange(rctx, code, oauth2.SetAuthURLParam("code_verifier", verifierState))
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err), nil
	}

	client := h.providerCfg.Client(rctx, token)
	resp, err := client.Get(h.userEndpointURL)
	if err != nil {
		return fmt.Errorf("userinfo request failed: %w", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("userinfo request returned non-200: code=%d", resp.StatusCode), nil
	}

	var user OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("userinfo decode failed: %w", err), nil
	}

	return nil, &user
}

// LoginCallback handles the OIDC code exchange and user info lookup.
func (h *oidcHandler) LoginCallback(res http.ResponseWriter, req *http.Request) LoginCallbackResult {
	logger := log.LoggerFromContext(req.Context())

	err, verifierState := h.verifyCookies(res, req)
	if err != nil {
		return LoginCallbackResult{ErrorMessage: err.Error()}
	}

	code := req.URL.Query().Get("code")
	if code == "" {
		return LoginCallbackResult{ErrorMessage: "missing code"}
	}

	err, user := h.exchangeUserinfo(req.Context(), code, verifierState)
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

	token, err := h.jwtH.sign(user.Name, int(pl))
	if err != nil {
		logger.Warn("failed to create a session", "err", err)
		return LoginCallbackResult{
			ErrorMessage: "failed to create a session",
		}
	}

	return LoginCallbackResult{Success: true,
		Token:           token,
		Username:        user.Name,
		PermissionLevel: pl,
	}
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

func getOIDCHandler() (AuthHandler, error) {
	cfg := config.GetConfig().Auth.OIDC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	discovery, err := fetchDiscovery(ctx, cfg.DiscoveryURL)

	if err != nil {
		return nil, err
	}

	provider := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  strings.TrimSuffix(config.GetConfig().Web.BaseURL, "/") + "/auth/code",
		Endpoint: oauth2.Endpoint{
			AuthURL:  discovery.AuthorizationEndpoint,
			TokenURL: discovery.TokenEndpoint,
		},
		Scopes: []string{"openid", "profile", "b3"},
	}

	return &oidcHandler{
		providerCfg:     provider,
		userEndpointURL: discovery.UserinfoEndpoint,
		jwtH: &jwtTokenHandler{
			secretKey: []byte(config.GetConfig().Web.SessionKey),
		},
	}, nil
}
