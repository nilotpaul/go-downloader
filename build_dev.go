//go:build dev
// +build dev

package main

import (
	"net/http"
)

func build() http.Handler {
	return http.NotFoundHandler()
}
