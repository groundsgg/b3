// SPDX-License-Identifier: AGPL-3.0-or-later
package core

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/request"
)

func Home(req *request.Request) {
	if req.OriginalRequest.URL.Path != "/" {
		req.PrintNotFound()
		return
	}

	req.OriginalWriter.WriteHeader(http.StatusOK)
	err := req.PrintOnly("pages.home", "Welcome to B3", nil)
	if err != nil {
		req.Logger.Error("failed to parse template",
			"template", "pages.home",
			"err", err,
		)
	}
}
