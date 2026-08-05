// Package config provides application configuration from environment variables.
package config

import (
	"errors"
	"os"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	SentryDSN   string
	DatabaseURL string
	BaseURL     string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		SentryDSN:   os.Getenv("SENTRY_DSN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		BaseURL:     os.Getenv("BASE_URL"),
	}

	if cfg.SentryDSN == "" {
		return nil, errors.New("SENTRY_DSN is not set")
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	if cfg.BaseURL == "" {
		return nil, errors.New("BASE_URL is not set")
	}

	return cfg, nil
}
