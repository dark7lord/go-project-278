// Package server wires the application together and runs its HTTP server.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"code/db/generated"
	"code/internal/adapters/httpapi"
	"code/internal/adapters/postgres"
	"code/internal/adapters/shortcode"
	"code/internal/application"
	"code/internal/config"
)

// connectDB creates a new pgxpool connection and pings the database.
func connectDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}

// setupRouter builds the gin engine: the middleware in order, then the routes.
func setupRouter(handler *httpapi.Handler) *gin.Engine {
	router := gin.New()
	router.Use(httpapi.RequestLog(slog.Default()))
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	router.Use(httpapi.RequestID())
	router.Use(httpapi.NewCORS())
	router.Use(gin.Recovery())

	router.TrustedPlatform = gin.PlatformCloudflare

	handler.RegisterRootRoutes(router)

	handler.RegisterAPIRoutes(router.Group("/api"))

	return router
}

// buildApp wires the stores, the service and the handler into a router.
func buildApp(cfg *config.Config, dbConn *pgxpool.Pool) *gin.Engine {
	queries := db.New(dbConn)
	linkRepo := postgres.NewLinkRepository(queries)
	visitRepo := postgres.NewVisitRepository(queries)
	service := application.NewService(linkRepo, visitRepo, shortcode.Generate)

	return setupRouter(httpapi.NewHandler(service, cfg.BaseURL))
}

const timeoutErrorBody = `{"error": "request timeout"}`

// withRequestTimeout answers a request over budget with a JSON 503. The content
// type goes on the outer writer: http.TimeoutHandler drops the inner headers.
func withRequestTimeout(h http.Handler, timeout time.Duration) http.Handler {
	timed := http.TimeoutHandler(h, timeout, timeoutErrorBody)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		timed.ServeHTTP(w, r)
	})
}

// maxRequestBodyBytes bounds request bodies (1 MiB headroom for JSON).
const maxRequestBodyBytes = 1 << 20

// writeTimeoutGrace lets the timeout 503 out before the write deadline, which
// would otherwise drop the connection.
const writeTimeoutGrace = 1 * time.Second

// newServer builds the HTTP server; its write deadline follows the request
// budget, so the timeout answer always gets out.
func newServer(cfg *config.Config, handler http.Handler) *http.Server {
	limited := http.MaxBytesHandler(handler, maxRequestBodyBytes)

	return &http.Server{
		Addr:              ":8080",
		Handler:           withRequestTimeout(limited, cfg.RequestTimeout),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      cfg.RequestTimeout + writeTimeoutGrace,
		IdleTimeout:       60 * time.Second,
	}
}

// Run loads config, connects to the database, and starts the HTTP server.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN}); err != nil {
		return fmt.Errorf("sentry.Init: %w", err)
	}
	defer sentry.Flush(2 * time.Second)

	ctx := context.Background()

	dbConn, err := connectDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbConn.Close()

	router := buildApp(cfg, dbConn)

	return serveUntilSignal(newServer(cfg, router))
}

// serveUntilSignal serves until the server fails or SIGINT/SIGTERM arrives;
// on a signal it drains in-flight requests and returns nil.
func serveUntilSignal(server *http.Server) error {
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("failed to run server: %w", err)

	case <-sigCtx.Done():
		slog.Info("signal received, shutting down", "timeout", "10s")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}

		slog.Info("in-flight requests drained, exiting")

		return nil
	}
}
