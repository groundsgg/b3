package b3

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed assets/**
var assetsFS embed.FS

// Assets returns a file system rooted at /assets for HTTP serving.
func Assets() (http.FileSystem, error) {
	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		return nil, fmt.Errorf("assets directory missing from embedded FS: %s", err.Error())
	}
	return http.FS(sub), nil
}
