// Package config provides application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
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

	baseURL, err := normalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	cfg.BaseURL = baseURL

	cfg.RequestTimeout = defaultRequestTimeout
	if raw := os.Getenv("REQUEST_TIMEOUT"); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("REQUEST_TIMEOUT: %w", err)
		}
		if timeout <= 0 {
			return nil, errors.New("REQUEST_TIMEOUT must be positive")
		}
		cfg.RequestTimeout = timeout
	}

	return cfg, nil
}

// normalizeBaseURL validates a public HTTP origin and removes its trailing slash.
func normalizeBaseURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("BASE_URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("BASE_URL must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("BASE_URL must include a host")
	}
	if parsed.User != nil {
		return "", errors.New("BASE_URL must not include user credentials")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", errors.New("BASE_URL must not include a path")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("BASE_URL must not include a query or fragment")
	}

	parsed.Path = ""

	return parsed.String(), nil
}
