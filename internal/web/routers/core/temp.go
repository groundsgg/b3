// SPDX-License-Identifier: AGPL-3.0-or-later
package core

import (
	"net/http"

	"github.com/groundsgg/b3/internal/web/request"
)

// A renders the pages.a template with a 200 response.
func A(req *request.Request) {
	req.OriginalWriter.WriteHeader(http.StatusOK)
	err := req.Print("pages.a", "AAA", nil)
	if err != nil {
		req.Logger.Error("failed to parse template",
			"template", "pages.a",
			"err", err,
		)
	}
}

// B renders the pages.b template with a 200 response.
func B(req *request.Request) {
	req.OriginalWriter.WriteHeader(http.StatusOK)
	err := req.Print("pages.b", "BBB", nil)
	if err != nil {
		req.Logger.Error("failed to parse template",
			"template", "pages.b",
			"err", err,
		)
	}
}
