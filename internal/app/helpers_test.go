package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/moby/moby/api/types/network"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"code/db/migrations"
	httpadapter "code/internal/adapters/http"
	postgresadapter "code/internal/adapters/postgres"
	"code/internal/application"
	"code/internal/db"
)

var testDBInst *testDB

func TestMain(m *testing.M) {
	ctx := context.Background()

	td, err := startTestDB(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	testDBInst = td

	code := m.Run()
	td.conn.Close()
	_ = td.pg.Terminate(ctx)
	os.Exit(code)
}

type testDB struct {
	pg      testcontainers.Container
	conn    *pgxpool.Pool
	queries *db.Queries
	repo    *postgresadapter.LinkRepository
	svc     *application.Service
	handler *httpadapter.Handler
	router  *gin.Engine
}

func startTestDB(ctx context.Context) (*testDB, error) {
	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategyAndDeadline(
			60*time.Second,
			wait.ForSQL("5432/tcp", "pgx", func(host string, port network.Port) string {
				return fmt.Sprintf("postgres://test:test@%s:%d/test?sslmode=disable", host, port.Num())
			}),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pg.Terminate(ctx)
		return nil, fmt.Errorf("get connection string: %w", err)
	}

	pool, err := connectDB(ctx, connStr)
	if err != nil {
		_ = pg.Terminate(ctx)
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	provider, err := goose.NewProvider("postgres", sqlDB, migrations.FS)
	if err != nil {
		pool.Close()
		_ = pg.Terminate(ctx)
		return nil, fmt.Errorf("create goose provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		pool.Close()
		_ = pg.Terminate(ctx)
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return newTestDB(pg, pool), nil
}

func newTestDB(pg testcontainers.Container, pool *pgxpool.Pool) *testDB {
	queries := db.New(pool)
	linkRepo := postgresadapter.NewLinkRepository(queries)
	linkSvc := application.NewServiceWithGenerator(application.ServiceDeps{
		LinkReader:    linkRepo,
		LinkWriter:    linkRepo,
		VisitReader:   linkRepo,
		VisitRecorder: linkRepo,
	}, "http://localhost:8080", stubGenerator{})
	linkHandler := httpadapter.NewHandler(linkSvc, linkSvc)
	router := setupRouter(linkHandler)

	return &testDB{
		pg:      pg,
		conn:    pool,
		queries: queries,
		repo:    linkRepo,
		svc:     linkSvc,
		handler: linkHandler,
		router:  router,
	}
}

func setupTestDB(t *testing.T) *testDB {
	t.Helper()
	return testDBInst
}

func setupTestTx(t *testing.T, td *testDB) *testDB {
	t.Helper()
	tx, err := td.conn.Begin(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })

	txQueries := td.queries.WithTx(tx)
	txRepo := postgresadapter.NewLinkRepository(txQueries)
	txSvc := application.NewServiceWithGenerator(application.ServiceDeps{
		LinkReader:    txRepo,
		LinkWriter:    txRepo,
		VisitReader:   txRepo,
		VisitRecorder: txRepo,
	}, "http://localhost:8080", stubGenerator{})
	txHandler := httpadapter.NewHandler(txSvc, txSvc)
	router := setupRouter(txHandler)

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
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), v))
}

func assertErrorBody(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]string
	decode(t, w, &body)
	assert.NotEmpty(t, body["error"])
}

func assertFieldErrors(t *testing.T, w *httptest.ResponseRecorder, field string) {
	t.Helper()
	var body map[string]map[string]string
	decode(t, w, &body)
	assert.NotEmpty(t, body["errors"][field])
}

type stubGenerator struct{}

func (stubGenerator) Generate() string {
	return "test-code"
}
