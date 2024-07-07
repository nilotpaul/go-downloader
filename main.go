package main

import (
	"log"

	"github.com/nilotpaul/go-downloader/api"
	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/store"
)

func main() {
	// Loads all Env vars from .env file.
	// Will panic if there's an error.
	env := config.MustLoadEnv()

	// Initializes and connects to DB.
	// Will panic if there's an error.
	db := config.MustInitDB(env.DBURL)
	defer db.Close()

	// Initializes the provider registry where all the
	// auth providers are registered.
	r := store.InitStore(*env, db)

	// Fiber API server
	s := api.NewAPIServer(env.Port, *env, r, db)

	// All routes, handlers & middlewares are registered
	// in func Start().
	log.Fatal(s.Start())
}
