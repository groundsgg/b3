// SPDX-License-Identifier: AGPL-3.0-or-later
package core

import (
	"net/http"

	"github.com/groundsgg/b3"
	"github.com/groundsgg/b3/internal/web/request"
)

// AssetsHandler serves embedded static assets under /assets/.
func AssetsHandler() (func(*request.Request), error) {
	fs, err := b3.Assets()
	if err != nil {
		return nil, err
	}
	handler := http.StripPrefix("/assets/", http.FileServer(fs))
	return func(r *request.Request) {
		handler.ServeHTTP(r.OriginalWriter, r.OriginalRequest)
	}, nil
}
