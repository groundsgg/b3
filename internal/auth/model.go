package auth

import (
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type AuthType string

const (
	BASIC_AUTH AuthType = "basic_auth"
	OIDC       AuthType = "oidc"
)

type LoginCallbackResult struct {
	Success         bool
	ErrorMessage    string
	Username        string
	PermissionLevel uint8
	Token           string
	SessionID       string
}

type AuthHandler interface {
	Type() AuthType
	PreLogin(http.ResponseWriter, *http.Request) error
	LoginCallback(http.ResponseWriter, *http.Request) LoginCallbackResult
	VerifyToken(string) (*UserInfo, error)
}

type UserInfo struct {
	Username        string
	PermissionLevel uint8
	NewToken        string
	SessionID       string
}

type basicUser struct {
	Name            string
	PasswordHash    []byte
	PermissionLevel uint8
}

type basicHandler struct {
	users map[string]*basicUser
	jwtH  *jwtTokenHandler
}

type OIDCDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
	Issuer                string `json:"issuer"`
}

type OIDCUserInfo struct {
	Subject string `json:"sub"`
	Name    string `json:"name"`
	B3Group string `json:"b3_group"`
}

type oidcHandler struct {
	oauthCfg *oauth2.Config
	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider
}

type oidcCookies struct {
	State    string
	Verifier string
	Nonce    string
}
