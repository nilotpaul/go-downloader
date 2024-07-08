package handler

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
)

type DownloadHandler struct {
	registry   *store.ProviderRegistry
	downloader *store.Downloader
	sessStore  *session.Store
}

func NewDownloadHandler(registry *store.ProviderRegistry, sessStore *session.Store) *DownloadHandler {
	return &DownloadHandler{
		registry:   registry,
		sessStore:  sessStore,
		downloader: store.NewDownloader(make([]string, 0), ""),
	}
}

func (h *DownloadHandler) DownloadHandler(c *fiber.Ctx) error {
	linksStr := c.Query("links", "")
	destPath := c.Query("path", "")
	if len(linksStr) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid link(s)",
		)
	}
	if len(destPath) == 0 {
		destPath = "./media"
	}

	fileIDs := util.GetFileIDs(linksStr)
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

	sess, err := util.GetSessionFromStore(c, h.sessStore)
	if err != nil {
		return util.NewAppError(
			http.StatusInternalServerError,
			"failed to get the session from store",
		)
	}

	h.downloader.FileIds = fileIDs
	h.downloader.DestinationPath = destPath
	h.downloader.UserID = sess.UserID

	slog.Info("downloading", "GDrive fileIDs: ", fileIDs)

	t := gp.GetAccessToken()
	if err := h.downloader.StartDownload(t, ""); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"fileID": fileIDs,
	})
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

	uID, ok := c.Locals(setting.LocalSessionKey).(string)
	if !ok {
		return util.NewAppError(
			websocket.TextMessage,
			"invalid UserID",
		)
	}
	for {
		pendings, _ := h.downloader.GetPendingDownloads(uID)
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

	return c.WriteMessage(websocket.TextMessage, []byte("no pending downloads"))
}
