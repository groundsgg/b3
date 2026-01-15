// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// GetAuthHandler builds the auth handler based on AUTH_TYPE and env config.
func GetAuthHandler(baseURL string) (AuthHandler, error) {
	at := os.Getenv("AUTH_TYPE")
	if at == "" {
		return nil, errors.New("env AUTH_TYPE is not defined")
	}

	switch strings.ToLower(at) {
	case "basic_auth":
		return getBasicAuthHandler()
	case "oidc":
		return getOIDCHandler(baseURL)
	}

	return nil, fmt.Errorf("unknown AUTH_TYPE '%s'", at)
}
