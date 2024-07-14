package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/nilotpaul/go-downloader/util"
)

type EnvConfig struct {
	Environment         string `envconfig:"ENVIRONMENT"`
	Port                string `envconfig:"PORT"`
	GoogleClientID      string `envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret  string `envconfig:"GOOGLE_CLIENT_SECRET"`
	AppURL              string `envconfig:"APP_URL"`
	DBURL               string `envconfig:"DB_URL"`
	SessionSecret       string `envconfig:"SESSION_SECRET"`
	Domain              string `envconfig:"DOMAIN"`
	DefaultDownloadPath string `envconfig:"DEFAULT_DOWNLOAD_PATH"`
}

func loadEnv() (*EnvConfig, error) {
	var cfg EnvConfig

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
	if err != nil {
		panic(err)
	}

	return cfg
}
