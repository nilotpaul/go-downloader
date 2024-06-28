package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/nilotpaul/go-downloader/service"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
)

type DownloadHandler struct {
	registry *store.ProviderRegistry
}

func NewDownloadHandler(registry *store.ProviderRegistry) *DownloadHandler {
	return &DownloadHandler{
		registry: registry,
	}
}

func (h *DownloadHandler) DownloadHandler(c *fiber.Ctx) error {
	fileID := c.Query("file_id", "")
	if len(fileID) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid file id",
		)
	}

	gp, err := h.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return err
	}

	t := gp.GetAccessToken()

	err = service.GDriveDownloader(fileID, "./media", "", t)
	if err != nil {
		return err
	}

	return c.JSON("download complete")
}
