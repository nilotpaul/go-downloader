package api

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/nilotpaul/go-downloader/api/handler"
	MW "github.com/nilotpaul/go-downloader/api/middleware"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
)

type Router struct {
	registry  *store.ProviderRegistry
	env       config.EnvConfig
	db        *sql.DB
	sessStore *session.Store
}

func NewRouter(registry *store.ProviderRegistry, env config.EnvConfig, db *sql.DB) *Router {
	return &Router{
		registry: registry,
		env:      env,
		db:       db,
	}
}

func (h *Router) RegisterRoutes(r fiber.Router) {
	// creates an in memory storage for session.
	store := session.New(session.Config{
		CookieDomain: h.env.Domain,
	})
	h.sessStore = store

	r.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.JSON("OK")
	})

	// Middlewares
	sessionMW := MW.NewSessionMiddleware(h.env, store, h.db, h.registry)

	// OAuth Handler for google.
	googleHR := handler.NewGoogleHandler(h.registry, store, h.db, h.env)
	r.Post("/signin/google", sessionMW.SessionMiddleware, sessionMW.WithoutGoogleOAuth, googleHR.GoogleSignInHandler)
	r.Post("/refresh", sessionMW.SessionMiddleware, sessionMW.WithGoogleOAuth, googleHR.RefreshTokenHandler)
	r.Post("/logout", sessionMW.WithGoogleOAuth, googleHR.LogoutHandler)
	r.Get("/callback/google", sessionMW.SessionMiddleware, sessionMW.WithoutGoogleOAuth, googleHR.GoogleCallbackHandler)

	downloadHR := handler.NewDownloadHandler(h.registry, h.sessStore, h.env)

	// For now this download route will only support GDrive, later multiple providers
	// will be handled here.
	r.Post("/download", sessionMW.SessionMiddleware, sessionMW.WithGoogleOAuth, downloadHR.DownloadHandler)
	r.Post("/cancel", sessionMW.SessionMiddleware, sessionMW.WithGoogleOAuth, downloadHR.CancelDownloadHandler)
	r.Post("/cancelAll", sessionMW.SessionMiddleware, sessionMW.WithGoogleOAuth, downloadHR.CancelAllDownloadsHandler)
	r.Get("/progress", sessionMW.SessionMiddleware, downloadHR.ProgressHTTPHandler)
	r.Get("/ws/progress", util.MakeWebsocketHandler(downloadHR.ProgressWebsocketHandler, h.env.AppURL))

	r.Get("/folderTree", downloadHR.FolderTreeHandler)
}
