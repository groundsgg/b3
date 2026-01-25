package auth

import (
	"net/http"
	"time"

	"github.com/groundsgg/b3/internal/config"
)

func setAuthCookie(res http.ResponseWriter, name, value string, lifetime time.Duration) {
	sameSite := http.SameSiteLaxMode
	secure := config.IsSecureConnection()
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(res, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		SameSite: sameSite,
		Secure:   secure,
		MaxAge:   int(lifetime.Seconds()),
		Path:     "/",
	})
}
