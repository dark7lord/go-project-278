package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://sentry.example.com/1")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("BASE_URL", "https://short.example.com")
	t.Setenv("REQUEST_TIMEOUT", "250ms")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "https://sentry.example.com/1", cfg.SentryDSN)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseURL)
	assert.Equal(t, "https://short.example.com", cfg.BaseURL)
	assert.Equal(t, 250*time.Millisecond, cfg.RequestTimeout)
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SENTRY_DSN", "")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("BASE_URL", "")
	t.Setenv("REQUEST_TIMEOUT", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Empty(t, cfg.SentryDSN)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseURL)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
	assert.Equal(t, defaultRequestTimeout, cfg.RequestTimeout)
}

func TestLoadMissingEnv(t *testing.T) {
	t.Setenv("SENTRY_DSN", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("BASE_URL", "")
	t.Setenv("REQUEST_TIMEOUT", "")

	_, err := Load()
	require.Error(t, err)
	assert.ErrorContains(t, err, "DATABASE_URL")
}

func TestLoadInvalidRequestTimeout(t *testing.T) {
	t.Setenv("SENTRY_DSN", "")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("BASE_URL", "")
	t.Setenv("REQUEST_TIMEOUT", "not-a-duration")

	_, err := Load()
	require.Error(t, err)
	assert.ErrorContains(t, err, "REQUEST_TIMEOUT")
}
