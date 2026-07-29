package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/moby/moby/api/types/network"
	"github.com/pressly/goose/v3"
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
	queries *db.Queries
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
	txRepo := repository.NewLinkRepository(txQueries)
	txSvc := service.NewLinkService(txRepo, "http://localhost:8080")
	txHandler := handler.NewLinkHandler(txSvc)
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
