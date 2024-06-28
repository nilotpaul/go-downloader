package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/util"
	"github.com/nilotpaul/go-downloader/www/page"
)

func HomeHandler(c *fiber.Ctx) error {
	return util.MakeTempl(c, page.Home())
}
