package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
)

type DownloadHandler struct {
	registry   *store.ProviderRegistry
	downloader *store.Downloader
	sessStore  *session.Store
	env        config.EnvConfig
}

func NewDownloadHandler(registry *store.ProviderRegistry, sessStore *session.Store, env config.EnvConfig) *DownloadHandler {
	return &DownloadHandler{
		registry:   registry,
		sessStore:  sessStore,
		env:        env,
		downloader: store.NewDownloader(make([]string, 0), ""),
	}
}

func (h *DownloadHandler) DownloadHandler(c *fiber.Ctx) error {
	b, err := util.ValidateDownloadHRBody(c)
	if err != nil {
		return err
	}
	if len(b.DestinationPath) == 0 {
		b.DestinationPath = h.env.DefaultDownloadPath
	}

	fileIDs := util.GetFileIDs(b.Links)
	if len(fileIDs) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid link(s)",
		)
	}

	gp, err := h.registry.GetProvider(setting.GoogleProvider)
	if err != nil {
		return util.NewAppError(
			http.StatusNotFound,
			"no provider found",
		)
	}

	t := gp.GetAccessToken()
	srv, err := util.MakeGDriveService(c.Context(), t)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to initialize the GDrive service",
			err,
		)
	}

	for _, folderID := range fileIDs["folder"] {
		Ids, err := util.GetFileIDsFromFolder(srv, folderID)
		if err != nil {
			return util.NewAppError(
				http.StatusInternalServerError,
				fmt.Sprintf("failed to get contents from folder %s. %s\n", folderID, err.Error()),
			)
		}
		fileIDs["file"] = append(fileIDs["file"], Ids...)
	}
	if len(fileIDs["file"]) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid link(s)",
		)
	}
	if util.HasDuplicates(fileIDs["file"]) {
		return util.NewAppError(
			http.StatusBadRequest,
			"duplicate links found",
		)
	}

	h.downloader.FileIds = fileIDs["file"]
	h.downloader.DestinationPath = b.DestinationPath

	slog.Info("downloading", "GDrive fileIDs: ", fileIDs)

	if err := h.downloader.StartDownload(c.Context(), t, ""); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"status":   http.StatusOK,
		"file_ids": fileIDs["file"],
	})
}

func (h *DownloadHandler) ProgressHTTPHandler(c *fiber.Ctx) error {
	pendings, _ := h.downloader.GetPendingDownloads()
	if len(pendings) == 0 {
		return util.NewAppError(
			http.StatusNotFound,
			"no pending downloads",
		)
	}

	return c.JSON(pendings)
}

func (h *DownloadHandler) CancelDownloadHandler(c *fiber.Ctx) error {
	fileID, err := util.ValidateCancelDownloadHRBody(c)
	if err != nil {
		return err
	}

	if _, err := h.downloader.GetProgress(fileID); err != nil {
		return util.NewAppError(
			http.StatusNotFound,
			"no ongoing downloads",
		)
	}

	if err := h.downloader.CancelDownload(fileID); err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			fmt.Sprintf("failed to cancel the download for file %s", fileID),
		)
	}

	return c.JSON("OK")
}

func (h *DownloadHandler) CancelAllDownloadsHandler(c *fiber.Ctx) error {
	h.downloader.CancelAllDownloads()

	return c.JSON("OK")
}

func (h *DownloadHandler) ProgressWebsocketHandler(c *websocket.Conn) error {
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("failed to close the ws connection: %v", err)
		}
	}()

	t := c.Cookies(setting.SessionKey, "")
	if len(t) == 0 {
		return util.NewAppError(
			websocket.TextMessage,
			"invalid session cookie",
		)
	}
	for {
		pendings, _ := h.downloader.GetPendingDownloads()
		progressJSON, err := json.Marshal(pendings)
		if err != nil {
			return util.NewAppError(
				websocket.TextMessage,
				"progress marshalling failed",
			)
		}

		if pendings != nil {
			if err := c.WriteMessage(websocket.TextMessage, progressJSON); err != nil {
				return err
			}
		}

		for fileID, errChan := range h.downloader.ErrChans {
			select {
			case err := <-errChan:
				errJSON, err := json.Marshal(fiber.Map{
					"file_id": fileID,
					"errMsg":  err.Error(),
				})
				if err != nil {
					if writeErr := c.WriteMessage(websocket.TextMessage, []byte("error marshalling failed")); writeErr != nil {
						return err
					}
				}
				if writeErr := c.WriteMessage(websocket.TextMessage, errJSON); writeErr != nil {
					return err
				}
			default:
			}
		}

		if len(pendings) == 0 {
			break
		}

		// Send progress updates in one and a half second interval
		time.Sleep(1500 * time.Millisecond)
	}

	infoMsg, err := json.Marshal(fiber.Map{
		"infoMsg": "no pending downloads",
	})
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.TextMessage, infoMsg)
}

func (h *DownloadHandler) FolderTreeHandler(c *fiber.Ctx) error {
	tree, err := util.GetFolderTree(".")
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to retrieve the folder tree",
		)
	}

	return c.JSON(tree)
}
