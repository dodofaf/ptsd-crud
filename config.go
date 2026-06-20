package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")

	err := viper.ReadInConfig()
	if err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{
		Port:        strings.TrimSpace(viper.GetString("PORT")),
		DatabaseURL: strings.TrimSpace(viper.GetString("DATABASE_URL")),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
