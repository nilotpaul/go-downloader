package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/util"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	status := http.StatusInternalServerError

	if apiErr, ok := err.(*util.AppError); ok {
		status = apiErr.Status

		slog.Error("HTTP API error", "errMsg", apiErr.Msg, "status", status, "err", apiErr.Err, "path", c.Path())

		return c.Status(status).JSON(fiber.Map{
			"status": status,
			"errMsg": apiErr.Msg,
		})
	}
	if fiberErr, ok := err.(*fiber.Error); ok {
		status = fiberErr.Code

		return c.Status(status).JSON(fiber.Map{
			"status": status,
			"errMsg": fiberErr.Message,
		})
	}

	slog.Error("HTTP API error", "err", err, "path", c.Path())

	return c.Status(status).JSON(fiber.Map{
		"status": status,
		"errMsg": "something went wrong",
	})
}
