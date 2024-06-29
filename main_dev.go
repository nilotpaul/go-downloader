// main_dev.go

//go:build dev
// +build dev

package main

import (
	"fmt"
	"net/http"
	"os"
)

func public() http.Handler {
	fmt.Println("building static files for development")
	return http.StripPrefix("/public/", http.FileServer(http.FS(os.DirFS("public"))))
}
