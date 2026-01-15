// SPDX-License-Identifier: AGPL-3.0-or-later
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

func setTokenCookie(w http.ResponseWriter, username string, permissionLevel uint8) {

}

func GetAuthHandler() (AuthHandler, error) {
	at := os.Getenv("AUTH_TYPE")
	if at == "" {
		return nil, errors.New("env AUTH_TYPE is not defined")
	}

	switch at {
	case "basic_auth":
		return getBasicAuthHandler()
	case "oauth2":
	}

	return nil, fmt.Errorf("unknown AUTH_TYPE '%s'", at)
}
