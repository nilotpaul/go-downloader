package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	MW "github.com/nilotpaul/go-downloader/api/middleware"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type publicFunc func() http.Handler

type APIServer struct {
	listenAddr string
	env        config.EnvConfig
	registry   *store.ProviderRegistry
	db         *sql.DB
	publicFunc publicFunc
}

func NewAPIServer(listenAddr string, env config.EnvConfig, registry *store.ProviderRegistry, db *sql.DB, publicFunc publicFunc) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		env:        env,
		registry:   registry,
		db:         db,
		publicFunc: publicFunc,
	}
}

func (s *APIServer) Start() error {
	app := fiber.New(fiber.Config{
		AppName:      "Go Downloader",
		ErrorHandler: MW.ErrorHandler,
	})
	logger := logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	})

	app.Use(logger)
	app.Use("/public/*", makeFiberHandler(s.publicFunc()))

	v1 := app.Group("/api/v1")

	handler := NewRouter(s.registry, s.env, s.db)
	handler.RegisterTemplRoutes(app)
	handler.RegisterRoutes(v1)

	log.Printf("Server started on http://localhost:%s", s.listenAddr)

	return app.Listen(":" + s.listenAddr)
}

func makeFiberHandler(h http.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(h)(c.Context())
		return nil
	}
}
