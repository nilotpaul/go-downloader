package store

import (
	"database/sql"
	"fmt"

	"github.com/nilotpaul/go-downloader/config"
	"github.com/nilotpaul/go-downloader/setting"
	"github.com/nilotpaul/go-downloader/types"
)

// `ProviderRegistry` holds all the cloud storage providers.
type ProviderRegistry struct {
	Providers map[string]types.OAuthProvider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		Providers: make(map[string]types.OAuthProvider),
	}
}

// Adds a provider in the `Providers` map.
func (r *ProviderRegistry) Register(providerName setting.Provider, p types.OAuthProvider) {
	provider := string(providerName)
	r.Providers[provider] = p
}

// Retrieves a provider from the `Providers` map.
func (r *ProviderRegistry) GetProvider(providerName setting.Provider) (types.OAuthProvider, error) {
	provider := string(providerName)
	p, exists := r.Providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider not found")
	}

	return p, nil
}

// `InitStore` initializes all the providers on start-up based on the provided env variables.
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
