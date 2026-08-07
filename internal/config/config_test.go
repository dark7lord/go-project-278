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

func TestLoadMissingEnv(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://sentry.example.com/1")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	t.Setenv("BASE_URL", "https://short.example.com")

	envVars := []string{"SENTRY_DSN", "DATABASE_URL", "BASE_URL"}

	for _, envVar := range envVars {
		t.Run("missing "+envVar, func(t *testing.T) {
			t.Setenv(envVar, "")

			_, err := Load()
			require.Error(t, err)
			assert.ErrorContains(t, err, envVar)
			assert.ErrorContains(t, err, "is not set")
		})
	}
}
