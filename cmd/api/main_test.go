package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/moby/moby/api/types/network"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"code/internal/db"
	"code/internal/handler"
	"code/internal/repository"
	"code/internal/service"
	"code/migrations"
)

type testDB struct {
	conn    *sql.DB
	repo    *repository.LinkRepository
	svc     *service.LinkService
	handler *handler.LinkHandler
	router  *gin.Engine
}

func setupTestDB(t *testing.T) *testDB {
	t.Helper()
	ctx := context.Background()

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategyAndDeadline(
			60*time.Second,
			wait.ForSQL("5432/tcp", "postgres", func(host string, port network.Port) string {
				return fmt.Sprintf("postgres://test:test@%s:%d/test?sslmode=disable", host, port.Num())
			}),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	conn, err := connectDB(connStr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	provider, err := goose.NewProvider("postgres", conn, migrations.FS)
	require.NoError(t, err)
	_, err = provider.Up(ctx)
	require.NoError(t, err)

	queries := db.New(conn)
	linkRepo := repository.NewLinkRepository(queries)
	linkSvc := service.NewLinkService(linkRepo, "http://localhost:8080")
	linkHandler := handler.NewLinkHandler(linkSvc)
	router := setupRouter(linkHandler)

	return &testDB{
		conn:    conn,
		repo:    linkRepo,
		svc:     linkSvc,
		handler: linkHandler,
		router:  router,
	}
}

func TestLinksCRUD(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateLink", func(t *testing.T) {
		body := `{"original_url": "https://example.com", "short_name": "my-link"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		td.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetLink / not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/links/999", nil)
		td.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GetLink", func(t *testing.T) {
		created, err := td.repo.CreateLink(ctx, "https://gettest.com", "get-test", "http://localhost:8080/get-test")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/links/%d", created.ID), nil)
		td.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp db.Link
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "https://gettest.com", resp.OriginalURL)
	})

	t.Run("UpdateLink", func(t *testing.T) {
		created, err := td.repo.CreateLink(ctx, "https://updatetest.com", "update-test", "http://localhost:8080/update-test")
		require.NoError(t, err)

		body := `{"original_url": "https://updated.com", "short_name": "updated-link"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/links/%d", created.ID), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		td.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteLink", func(t *testing.T) {
		created, err := td.repo.CreateLink(ctx, "https://deletetest.com", "delete-test", "http://localhost:8080/delete-test")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/links/%d", created.ID), nil)
		td.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("ListLinks", func(t *testing.T) {
		_, err := td.repo.CreateLink(ctx, "https://list1.com", "list-link-1", "http://localhost:8080/list-link-1")
		require.NoError(t, err)
		_, err = td.repo.CreateLink(ctx, "https://list2.com", "list-link-2", "http://localhost:8080/list-link-2")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/links", nil)
		td.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		assert.NotEqual(t, "null", w.Body.String())
	})
}
