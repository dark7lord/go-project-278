// Package app provides application initialization and HTTP server setup.
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "code/internal/adapters/http"
	"code/internal/adapters/postgres"
	shortcodeadapter "code/internal/adapters/shortcode"
	"code/internal/application"
	"code/internal/config"
	"code/internal/db"
)

func init() {
	binding.EnableDecoderDisallowUnknownFields = true
}

// Check that the httpadapter middleware wiring stays transport-local: the
// composition root no longer defines HTTP cross-cutting concerns here.

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
	router.Use(gin.Logger())
	router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	router.Use(httpadapter.NewCORS())
	router.Use(gin.Recovery())

	router.TrustedPlatform = gin.PlatformCloudflare

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.GET("/r/:code", linkHandler.Redirect)

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

	server := &http.Server{
		Addr:              ":8080",
		Handler:           http.TimeoutHandler(router, cfg.RequestTimeout, `{"error": "request timeout"}`),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return serveUntilSignal(server)
}

// serveUntilSignal serves HTTP requests until the server crashes on its own
// or a shutdown signal (SIGINT/SIGTERM) arrives. On a signal it drains
// in-flight requests via http.Server.Shutdown and returns nil.
func serveUntilSignal(server *http.Server) error {
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("[server] listening on %s", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("failed to run server: %w", err)

	case <-sigCtx.Done():
		log.Printf("[server] signal received, shutting down gracefully (timeout 10s)...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown: %w", err)
		}

		log.Printf("[server] drained in-flight requests, exiting")

		return nil
	}
}
