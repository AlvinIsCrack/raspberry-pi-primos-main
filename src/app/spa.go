package app

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

// setupSPAFallback monta el sistema de archivos embebido en el mux con soporte SPA.
func setupSPAFallback(mux *http.ServeMux, assets fs.FS, assetsDir string) error {
	buildSubtree, err := fs.Sub(assets, assetsDir)
	if err != nil {
		return fmt.Errorf("static assets initialization failed: %w", err)
	}

	fileServer := http.FileServer(http.FS(buildSubtree))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		stat, err := fs.Stat(buildSubtree, path)
		if errors.Is(err, fs.ErrNotExist) || (err == nil && stat.IsDir()) {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	return nil
}
