//go:build !dev
// +build !dev

package main

import (
	"net/http"
	"os"
	"path/filepath"
)

func build() http.Handler {
	distDir := "./dist"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the absolute path to prevent directory traversal
		absPath, err := filepath.Abs(filepath.Join(distDir, r.URL.Path))
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Check if the file exists
		if _, err := os.Stat(absPath); err != nil {
			// Serve the index.html file
			http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
			return
		}

		// Otherwise, serve the static file
		http.FileServer(http.Dir(distDir)).ServeHTTP(w, r)
	})
}
