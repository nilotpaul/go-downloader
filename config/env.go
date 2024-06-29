package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type EnvConfig struct {
	Environment        string `envconfig:"ENVIRONMENT"`
	Port               string `envconfig:"PORT"`
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
	AppURL             string `envconfig:"APP_URL"`
	DBURL              string `envconfig:"DB_URL"`
	SessionSecret      string `envconfig:"SESSION_SECRET"`
}

func loadEnv() (*EnvConfig, error) {
	var cfg EnvConfig

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	err = envconfig.Process("", &cfg)
	if err != nil {
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
