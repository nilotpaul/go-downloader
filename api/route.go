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
	store := session.New()
	h.sessStore = store

	r.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.JSON("OK")
	})

	// Middlewares
	sessionMW := MW.NewSessionMiddleware(h.env, store, h.db, h.registry)
	r.Use(sessionMW.SessionMiddleware)

	// OAuth Handler for google.
	googleHR := handler.NewGoogleHandler(h.registry, store, h.db)
	r.Get("/signin/google", sessionMW.WithoutGoogleOAuth, googleHR.GoogleSignInHandler)
	r.Get("/callback/google", sessionMW.WithoutGoogleOAuth, googleHR.GoogleCallbackHandler)
	r.Get("/session",
		sessionMW.WithGoogleOAuth,
		googleHR.GetSessionHandler,
	)
	r.Post("/logout", sessionMW.WithGoogleOAuth, googleHR.LogoutHandler)

	downloadHR := handler.NewDownloadHandler(h.registry, h.sessStore)
	r.Get("/download",
		sessionMW.WithGoogleOAuth,
		downloadHR.DownloadHandler,
	)
	r.Get("/ws/progress",
		sessionMW.WithGoogleOAuth,
		util.MakeWebsocketHandler(downloadHR.ProgressWebsocketHandler),
	)

	// just for testing
	r.Get("/test",
		sessionMW.WithGoogleOAuth,
		func(c *fiber.Ctx) error {
			s, _ := util.GetSessionFromStore(c, store)
			return c.JSON(s)
		})

}
