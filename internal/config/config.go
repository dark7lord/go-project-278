// Package config provides application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	SentryDSN      string
	DatabaseURL    string
	BaseURL        string
	RequestTimeout time.Duration
}

// defaultBaseURL is used when BASE_URL is not set.
const defaultBaseURL = "http://localhost:8080"

// defaultRequestTimeout bounds the execution time of a single HTTP request.
const defaultRequestTimeout = 10 * time.Second

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

	cfg.RequestTimeout = defaultRequestTimeout
	if raw := os.Getenv("REQUEST_TIMEOUT"); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("REQUEST_TIMEOUT: %w", err)
		}
		cfg.RequestTimeout = timeout
	}

	return cfg, nil
}
