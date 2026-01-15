// SPDX-License-Identifier: AGPL-3.0-or-later
package request

import (
	"net/http"

	"github.com/groundsgg/b3/internal/auth"
	"github.com/groundsgg/b3/internal/web/pages"
)

func Parse(f func(req *Request), pages pages.Pages, ah auth.AuthHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := &Request{
			OriginalWriter:  w,
			OriginalRequest: r,
			pages:           pages,
			AuthHandler:     ah,
		}
		f(req)
	})
}
