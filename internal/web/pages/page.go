// SPDX-License-Identifier: AGPL-3.0-or-later
package pages

import (
	"net/http"
)

type PageData struct {
	Title     string
	Session   any
	Data      any
	RequestID string
}

type Pages interface {
	Render(w http.ResponseWriter, name string, data PageData) error
}
