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

// Print renders a page template with the provided title and data.
// No status code is set.
func (r *Request) Print(pageID, title string, data any) error {
	return r.pages.Render(r.OriginalWriter, pageID, pages.PageData{
		Title:   title,
		Data:    data,
		Session: r.Session,
	})
}

// PrintNotFound renders a 404 error page.
func (r *Request) PrintNotFound() {
	r.PrintError(ErrorData{
		Code:    http.StatusNotFound,
		Message: "ressource not found",
	})
}

// PrintError writes an HTTP error code and renders the error page.
func (r *Request) PrintError(data ErrorData) {
	r.OriginalWriter.WriteHeader(data.Code)
	r.Print("pages.error", "Error", data)
}
