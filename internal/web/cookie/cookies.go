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
)

func Delete(res http.ResponseWriter, name string) {
	Set(res, name, "", -1)
}

func Set(res http.ResponseWriter, name, value string, lifetime time.Duration) {
	sameSite := http.SameSiteLaxMode
	secure := config.IsSecureConnection()
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	c := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   secure,
		MaxAge:   int(lifetime.Seconds()),
		Path:     "/",
	}

	if lifetime == -1 {
		c.MaxAge = -1
	}

	http.SetCookie(res, c)
}
