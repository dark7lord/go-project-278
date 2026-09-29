// Package app provides application initialization and HTTP server setup.
package app

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
	"code/db/migrations"
	httpadapter "code/internal/adapters/http"
	"code/internal/adapters/postgres"
	shortcodeadapter "code/internal/adapters/shortcode"
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

// setupRouter creates and configures the gin engine with all routes.
func setupRouter(linkHandler *httpadapter.Handler) *gin.Engine {
	router := gin.New()
	router.Use(httpadapter.RequestLog(slog.Default()))
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	router.Use(httpadapter.RequestID())
	router.Use(httpadapter.NewCORS())
	router.Use(gin.Recovery())

	router.TrustedPlatform = gin.PlatformCloudflare

	linkHandler.RegisterRootRoutes(router)

	api := router.Group("/api")
	api.Use(httpadapter.MaxRequestBody(httpadapter.MaxRequestBodyBytes))
	linkHandler.RegisterAPIRoutes(api)

	return router
}

func newService(
	linkRepo *postgres.LinkRepository,
	visitRepo *postgres.VisitRepository,
) *application.Service {
	shortCodeGenerator := shortcodeadapter.NewGenerator()

	return application.NewServiceWithGenerator(application.ServiceDeps{
		LinkReader:    linkRepo,
		LinkWriter:    linkRepo,
		VisitReader:   visitRepo,
		VisitRecorder: visitRepo,
	}, shortCodeGenerator)
}

func newLinkHandler(
	linkService application.LinkUseCase,
	visitService application.VisitUseCase,
	baseURL string,
) *httpadapter.Handler {
	return httpadapter.NewHandler(linkService, visitService, baseURL)
}

// buildApp assembles the application dependencies and HTTP router in one explicit composition root.
func buildApp(cfg *config.Config, dbConn *pgxpool.Pool) *gin.Engine {
	queries := db.New(dbConn)
	linkRepo := postgres.NewLinkRepository(queries)
	visitRepo := postgres.NewVisitRepository(queries)
	linkService := newService(linkRepo, visitRepo)
	linkHandler := newLinkHandler(linkService, linkService, cfg.BaseURL)

	return setupRouter(linkHandler)
}

const timeoutErrorBody = `{"error": "request timeout"}`

// withRequestTimeout bounds a request and answers an expired one in JSON.
// The content type is set on the outer writer on purpose: http.TimeoutHandler
// discards the handler's own headers on the timeout path, so a type set by the
// handler or by gin would never reach the client.
func withRequestTimeout(h http.Handler, timeout time.Duration) http.Handler {
	timed := http.TimeoutHandler(h, timeout, timeoutErrorBody)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		timed.ServeHTTP(w, r)
	})
}

// writeTimeoutGrace is the window the timeout handler gets to deliver its
// response after the request budget is spent. Without it the socket deadline
// races the middleware's 503 and the client sees a dropped connection instead.
const writeTimeoutGrace = 1 * time.Second

// newServer builds the HTTP server. The write deadline is derived from the
// request budget so it can never preempt the timeout handler's response.
func newServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":8080",
		Handler:           withRequestTimeout(handler, cfg.RequestTimeout),
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

	if err := migrations.Up(ctx, dbConn); err != nil {
		return err
	}

	router := buildApp(cfg, dbConn)

	return serveUntilSignal(newServer(cfg, router))
}

// serveUntilSignal serves HTTP requests until the server crashes on its own
// or a shutdown signal (SIGINT/SIGTERM) arrives. On a signal it drains
// in-flight requests via http.Server.Shutdown and returns nil.
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
