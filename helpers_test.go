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

	"code/db"
	"code/links"
	"code/migrations"
)

type testDB struct {
	conn    *sql.DB
	queries *db.Queries
	repo    *links.Repository
	svc     *links.Service
	handler *links.Handler
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
	linkRepo := links.NewRepository(queries)
	linkSvc := links.NewService(linkRepo, "http://localhost:8080")
	linkHandler := links.NewHandler(linkSvc)
	router := setupRouter(linkHandler, "")

	return &testDB{
		conn:    conn,
		queries: queries,
		repo:    linkRepo,
		svc:     linkSvc,
		handler: linkHandler,
		router:  router,
	}
}

func setupTestTx(t *testing.T, td *testDB) *testDB {
	t.Helper()
	tx, err := td.conn.Begin()
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	txQueries := td.queries.WithTx(tx)
	txRepo := links.NewRepository(txQueries)
	txSvc := links.NewService(txRepo, "http://localhost:8080")
	txHandler := links.NewHandler(txSvc)
	router := setupRouter(txHandler, "")

	return &testDB{
		conn:    td.conn,
		queries: txQueries,
		repo:    txRepo,
		svc:     txSvc,
		handler: txHandler,
		router:  router,
	}
}

func linkFactory(i int) db.Link {
	originalURL := fmt.Sprintf("https://link-%d.com", i)
	shortName := fmt.Sprintf("%d-%d-%d", i, i, i)
	shortURL := fmt.Sprintf("%s/%s", originalURL, shortName)

	return db.Link{
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	}
}

func performRequest(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	r.ServeHTTP(w, req)

	return w
}

func assertErrorBody(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.NotEmpty(t, body["error"])
}
