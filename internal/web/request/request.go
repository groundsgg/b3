// SPDX-License-Identifier: AGPL-3.0-or-later
package request

import (
	"log/slog"
	"net/http"

	"github.com/groundsgg/b3/internal/auth"
	"github.com/groundsgg/b3/internal/web/pages"
)

type Request struct {
	OriginalWriter  http.ResponseWriter
	OriginalRequest *http.Request
	Logger          *slog.Logger
	ID              string
	pages           pages.Pages
	Session         SessionInfo
	AuthHandler     auth.AuthHandler
}

type ErrorData struct {
	Message string
	Code    int
}

func (r *Request) PrintOnly(pageID, title string, data any) error {
	return r.pages.Render(r.OriginalWriter, pageID, pages.PageData{
		Title:   title,
		Data:    data,
		Session: r.Session,
	})
}

func (r *Request) PrintNotFound() {
	r.PrintError(ErrorData{
		Code:    http.StatusNotFound,
		Message: "ressource not found",
	})
}

func (r *Request) PrintError(data ErrorData) {
	r.OriginalWriter.WriteHeader(data.Code)
	r.PrintOnly("pages.error", "Error", data)
}
