package handler

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strings"
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
	d := store.NewDownloader(make([]string, 0), "")

	return &DownloadHandler{
		registry:   registry,
		downloader: d,
		sessStore:  sessStore,
	}
}

func (h *DownloadHandler) DownloadHandler(c *fiber.Ctx) error {
	link := c.Query("link", "")
	if len(link) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid link",
		)
	}
	if ok := strings.Contains(link, ","); ok {
		return util.NewAppError(
			http.StatusNotImplemented,
			"multiple links are not currently supported",
		)
	}
	fileID := util.GetGDriveFileID(link)
	if len(fileID) == 0 {
		return util.NewAppError(
			http.StatusBadRequest,
			"invalid link",
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

	h.downloader.FileIds = []string{fileID}
	h.downloader.DestinationPath = "./media"
	h.downloader.UserID = sess.UserID

	slog.Info("downloading", "GDriveURL", link, "fileID", fileID)

	t := gp.GetAccessToken()
	if err := h.downloader.StartDownload(t, ""); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"fileID": fileID,
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

	fileID := c.Query("fileID", "")
	if len(fileID) == 0 {
		return util.NewAppError(
			websocket.TextMessage,
			"invalid fileID",
		)
	}

	uID, ok := c.Locals(setting.LocalSessionKey).(string)
	if !ok {
		return util.NewAppError(
			websocket.TextMessage,
			"invalid UserID",
		)
	}

	infoMsg := []byte("you can only receive updates, cannot send messages")
	if err := c.WriteMessage(websocket.TextMessage, infoMsg); err != nil {
		log.Println("Error sending info message:", err)
	}

	for {
		prog, err := h.downloader.GetProgress(fileID, uID)
		if err != nil {
			return util.NewAppError(
				websocket.TextMessage,
				"no ongoing downloads found with this fileID",
			)
		}

		progressJSON, err := json.Marshal(prog)
		if err != nil {
			return util.NewAppError(
				websocket.TextMessage,
				"progress marshalling failed",
			)
		}

		if err := c.WriteMessage(websocket.TextMessage, progressJSON); err != nil {
			return err
		}

		if prog.Complete {
			break
		}

		// Send progress updates in one and a half second interval
		time.Sleep(1500 * time.Millisecond)
	}

	return c.WriteMessage(websocket.TextMessage, []byte("download completed"))
}
