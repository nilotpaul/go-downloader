//go:build !dev
// +build !dev

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed dist/*
var dist embed.FS

func build() fiber.Handler {
	dist, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Fatalf("Failed to get sub filesystem: %v", err)
	}

	fileServer := filesystem.New(filesystem.Config{
		Root:   http.FS(dist),
		Browse: true,
	})

	return func(c *fiber.Ctx) error {
		path := strings.TrimPrefix(c.Path(), "/")
		_, err := fs.Stat(dist, path)
		if os.IsNotExist(err) {
			// If the requested file doesn't exist, serve index.html.
			// These paths will be handled by client-side router.
			fmt.Println("serving index.html")
			return c.SendFile("dist/index.html")
		}
		return fileServer(c)
	}
}
