//go:build dev
// +build dev

package main

import "github.com/gofiber/fiber/v2"

func build() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}
