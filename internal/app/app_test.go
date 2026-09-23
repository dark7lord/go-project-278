package app

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpadapter "code/internal/adapters/http"
)

func TestConnectDBInvalidDSN(t *testing.T) {
	_, err := connectDB(context.Background(), "not-a-valid-dsn")
	require.Error(t, err)
}

func TestConnectDBUnreachable(t *testing.T) {
	_, err := connectDB(context.Background(), "postgres://user:pass@127.0.0.1:1/db?sslmode=disable")
	require.Error(t, err)
}

func TestPingRoute(t *testing.T) {
	router := setupRouter(httpadapter.NewHandler(nil, nil, "http://localhost:8080"))

	w := performRequest(t, router, "GET", "/ping", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestRequestBodyTooLarge(t *testing.T) {
	router := setupRouter(httpadapter.NewHandler(nil, nil, "http://localhost:8080"))

	body := strings.Repeat("a", maxRequestBodyBytes+1)
	w := performRequest(t, router, "POST", "/api/links", body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assertErrorBody(t, w)
}
