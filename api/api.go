package api

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	MW "github.com/nilotpaul/go-downloader/api/middleware"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/store"
)

type buildFunc func() fiber.Handler

type APIServer struct {
	listenAddr string
	env        config.EnvConfig
	registry   *store.ProviderRegistry
	db         *sql.DB
	build      buildFunc
}

func NewAPIServer(listenAddr string, env config.EnvConfig, registry *store.ProviderRegistry, db *sql.DB, build buildFunc) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		env:        env,
		registry:   registry,
		db:         db,
		build:      build,
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

	// Middlewares
	app.Use(logger)
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     fmt.Sprintf("http://localhost:5173,%s", s.env.AppURL),
		AllowMethods:     "GET,POST",
	}))

	// API Routes will be prefixed with `/api/v1`.
	v1 := app.Group("/api/v1")

	r := NewRouter(s.registry, s.env, s.db)
	r.RegisterRoutes(v1)

	// Static build folder for production usage.
	// In development, this won't do anything.
	// In production, for all non API routes, it'll serve the static `dist` dir.
	app.Use("/", s.build())

	log.Printf("Server started on http://localhost:%s", s.listenAddr)

	return app.Listen(":" + s.listenAddr)
}
