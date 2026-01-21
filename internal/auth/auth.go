package auth

import (
	"fmt"

	"github.com/groundsgg/b3/internal/config"
)

// GetAuthHandler builds the auth handler based on AUTH_TYPE and env config.
func GetAuthHandler() (AuthHandler, error) {
	switch config.GetConfig().Auth.Type {
	case "basic_auth":
		return getBasicAuthHandler()
	case "oidc":
		return getOIDCHandler()
	}

	return nil, fmt.Errorf("unknown AUTH_TYPE '%s'", config.GetConfig().Auth.Type)
}
