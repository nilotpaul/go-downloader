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
	registry *store.ProviderRegistry
	env      config.EnvConfig
	db       *sql.DB
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

	r.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.JSON("OK")
	})

	// Middlewares
	sessionMW := MW.NewSessionMiddleware(h.env, store, h.db, h.registry)

	// OAuth Handler for google.
	googleHR := handler.NewGoogleHandler(h.registry)
	r.Get("/signin/google", sessionMW.WithoutGoogleOAuth, googleHR.GoogleSignInHandler)
	r.Post("/callback/google", sessionMW.WithoutGoogleOAuth, googleHR.GoogleCallbackHandler)
	r.Post("/logout", sessionMW.WithGoogleOAuth, googleHR.LogoutHandler)

	downloadHR := handler.NewDownloadHandler(h.registry)
	r.Get("/download",
		sessionMW.SessionMiddleware,
		sessionMW.WithGoogleOAuth,
		downloadHR.DownloadHandler,
	)

	// just for testing
	r.Get("/test",
		sessionMW.SessionMiddleware,
		sessionMW.WithGoogleOAuth,
		func(c *fiber.Ctx) error {
			s, _ := util.GetSessionFromStore(c, store)
			return c.JSON(s)
		})
}

func (h *Router) RegisterTemplRoutes(r fiber.Router) {
	r.Get("/", handler.HomeHandler)
}
