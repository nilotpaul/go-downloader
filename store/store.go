package store

import (
	"database/sql"

	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/setting"
)

func InitStore(env config.EnvConfig, db *sql.DB) *ProviderRegistry {
	r := NewProviderRegistry()

	if len(env.GoogleClientSecret) != 0 || len(env.GoogleClientID) != 0 {
		googleProvider := NewGoogleProvider(googleProviderConfig{
			googleClientID:     env.GoogleClientID,
			googleClientSecret: env.GoogleClientSecret,
			googleRedirectURL:  env.AppURL + "/api/v1/callback/google",
		}, db, env)

		r.Register(setting.GoogleProvider, googleProvider)
	}

	return r
}
