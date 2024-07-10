package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	MW "github.com/nilotpaul/go-downloader/api/middleware"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type APIServer struct {
	listenAddr string
	env        config.EnvConfig
	registry   *store.ProviderRegistry
	db         *sql.DB
}

func NewAPIServer(listenAddr string, env config.EnvConfig, registry *store.ProviderRegistry, db *sql.DB) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		env:        env,
		registry:   registry,
		db:         db,
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
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST",
	}))

	v1 := app.Group("/api/v1")

	handler := NewRouter(s.registry, s.env, s.db)
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
