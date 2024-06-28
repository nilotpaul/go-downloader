// main_non_dev.go

//go:build !dev
// +build !dev

package main

import (
	"embed"
	"fmt"
	"net/http"
)

//go:embed public
var publicFS embed.FS

func public() http.Handler {
	fmt.Println("building static files for production")
	return http.FileServer(http.FS(publicFS))
}
