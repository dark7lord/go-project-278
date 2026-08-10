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

// defaultBaseURL is used when BASE_URL is not set.
const defaultBaseURL = "http://localhost:8080"

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		SentryDSN:   os.Getenv("SENTRY_DSN"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		BaseURL:     os.Getenv("BASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}

	return cfg, nil
}
