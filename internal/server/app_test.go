package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpadapter "code/internal/adapters/http"
)

func TestConnectDBInvalidDSN(t *testing.T) {
	_, err := connectDB(t.Context(), "not-a-valid-dsn")
	require.Error(t, err)
}

func TestConnectDBUnreachable(t *testing.T) {
	_, err := connectDB(t.Context(), "postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.Error(t, err)
}

func TestPingRoute(t *testing.T) {
	router := setupRouter(httpadapter.NewHandler(nil, "http://localhost:8080"))

	w := performRequest(t, router, "GET", "/ping", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
