// SPDX-License-Identifier: AGPL-3.0-or-later
package b3

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets/**
var assetsFS embed.FS

// Assets returns a file system rooted at /assets for HTTP serving.
func Assets() http.FileSystem {
	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		panic("assets directory missing from embedded FS")
	}
	return http.FS(sub)
}
