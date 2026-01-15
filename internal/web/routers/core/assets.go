// SPDX-License-Identifier: AGPL-3.0-or-later
package core

import (
	"net/http"

	"github.com/groundsgg/b3"
	"github.com/groundsgg/b3/internal/web/request"
)

func AssetsHandler() func(*request.Request) {
	handler := http.StripPrefix("/assets/", http.FileServer(b3.Assets()))
	return func(r *request.Request) {
		handler.ServeHTTP(r.OriginalWriter, r.OriginalRequest)
	}
}
