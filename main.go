package main

import (
	"log"

	"github.com/nilotpaul/go-downloader/api"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/store"
	"github.com/nilotpaul/go-downloader/util"
	"github.com/pressly/goose/v3"
)

func main() {
	// Loads all Env vars from .env file.
	env := config.MustLoadEnv()

	// Initializes and connects to DB.
	db := config.MustInitDB(env.DBURL)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing the database connection: %s", err)
		}
	}()

	// Run auto migrations in production.
	if util.IsProduction() {
		if err := goose.Up(db, "./migrations"); err != nil {
			log.Fatalf("failed to apply database migrations: %v", err)
		}
	}

	// Initializes the provider registry where all the
	// auth providers are registered.
	r := store.InitStore(*env, db)

	s := api.NewAPIServer(env.Port, *env, r, db, build)

	// All routes, handlers & middlewares are registered
	// in func Start().
	log.Fatal(s.Start())
}
