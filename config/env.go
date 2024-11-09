package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/nilotpaul/go-downloader/util"
)

// Main env configuration
type EnvConfig struct {
	Environment         string `envconfig:"ENVIRONMENT" default:"DEV"`
	Port                string `envconfig:"PORT" required:"true"`
	DBURL               string `envconfig:"DB_URL" required:"true"`
	AppURL              string `envconfig:"APP_URL" required:"true"`
	Domain              string `envconfig:"DOMAIN" required:"true"`
	DefaultDownloadPath string `envconfig:"DEFAULT_DOWNLOAD_PATH"`

	SessionSecret string `envconfig:"SESSION_SECRET" required:"true"`
	GoogleOAuthEnvConfig
}

// Google OAuth specific configuration
type GoogleOAuthEnvConfig struct {
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID" required:"true"`
	GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET" required:"true"`
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
