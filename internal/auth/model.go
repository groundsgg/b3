// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import "net/http"

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
}

type AuthHandler interface {
	Type() AuthType
	PreLogin(http.ResponseWriter, *http.Request) error
	LoginCallback(http.ResponseWriter, *http.Request) LoginCallbackResult
}

type OIDCDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
	Issuer                string `json:"issuer"`
}

type OIDCUserInfo struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	B3Group    string `json:"b3_group"`
}
