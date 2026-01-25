package auth

import (
	"net/http"
	"time"

	"github.com/groundsgg/b3/internal/config"
)

func setCookie(res http.ResponseWriter, name, value string, lifetime time.Duration) {
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
