// SPDX-License-Identifier: AGPL-3.0-or-later
package b3

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/groundsgg/b3/internal/web/pages"
)

//go:embed templates/**/*.html
var templateFS embed.FS

type Renderer struct {
	t *template.Template
}

// Render executes the named template with the provided page data.
func (r *Renderer) Render(w http.ResponseWriter, name string, data pages.PageData) error {
	return r.t.ExecuteTemplate(w, name, data)
}

// NewRenderer parses embedded HTML templates and returns a renderer.
func NewRenderer() (*Renderer, error) {
	t, err := template.New("").
		ParseFS(templateFS, "templates/**/*.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{t: t}, nil
}
