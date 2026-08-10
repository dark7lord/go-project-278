package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://sentry.example.com/1")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("BASE_URL", "https://short.example.com")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "https://sentry.example.com/1", cfg.SentryDSN)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseURL)
	assert.Equal(t, "https://short.example.com", cfg.BaseURL)
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SENTRY_DSN", "")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("BASE_URL", "")

	cfg, err := Load()
	require.NoError(t, err)

	assert.Empty(t, cfg.SentryDSN)
	assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DatabaseURL)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
}

func TestLoadMissingEnv(t *testing.T) {
	t.Setenv("SENTRY_DSN", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("BASE_URL", "")

	_, err := Load()
	require.Error(t, err)
	assert.ErrorContains(t, err, "DATABASE_URL")
}
