package cookie

import (
	"net/http"
	"time"

	"github.com/groundsgg/b3/internal/config"
)

const (
	OIDC_STATE    string = "oidc_state"
	OIDC_VERIFIER string = "oidc_verifier"
	OIDC_NONCE    string = "oidc_nonce"
	AUTH_TOKEN    string = "auth_token"

	AUTH_TOKEN_LIFETIME time.Duration = time.Hour * 24
)

// Delete clears a cookie by setting it with a negative lifetime.
func Delete(res http.ResponseWriter, name string) {
	Set(res, name, "", -1)
}

// Set sets an HTTP-only cookie with the given name/value and lifetime.
// Cookie security attributes are derived from the configured base URL.
func Set(res http.ResponseWriter, name, value string, lifetime time.Duration) {
	secure := config.IsSecureConnection()

	c := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int(lifetime.Seconds()),
		Path:     "/",
	}

	if lifetime == -1 {
		c.MaxAge = -1
	}

	http.SetCookie(res, c)
}
