package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/nilotpaul/go-downloader/util"
)

// Main env configuration
type EnvConfig struct {
	Environment         string `envconfig:"ENVIRONMENT"`
	Port                string `envconfig:"PORT"`
	DBURL               string `envconfig:"DB_URL"`
	AppURL              string `envconfig:"APP_URL"`
	Domain              string `envconfig:"DOMAIN"`
	DefaultDownloadPath string `envconfig:"DEFAULT_DOWNLOAD_PATH"`

	SessionSecret string `envconfig:"SESSION_SECRET"`
	GoogleOAuthEnvConfig
}

// Google OAuth specific configuration
type GoogleOAuthEnvConfig struct {
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
}

func loadEnv() (*EnvConfig, error) {
	var cfg EnvConfig

	// Read env vars from `.env` file in development.
	// In production, it'll default to the runtime injected variables.
	if !util.IsProduction() {
		if err := godotenv.Load(); err != nil {
			return nil, err
		}
	}
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustLoadEnv() *EnvConfig {
	cfg, err := loadEnv()

	// All required env needs to be loaded, else we panic.
	if err != nil {
		panic(err)
	}

	return cfg
}
