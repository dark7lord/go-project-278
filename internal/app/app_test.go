package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/db"
	"code/internal/link"
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
	repo := link.NewRepository(db.New(nil))
	svc := link.NewService(repo, "http://localhost:8080")
	router := setupRouter(link.NewHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"pong"}`, w.Body.String())
}
