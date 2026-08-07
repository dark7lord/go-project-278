// Package app provides application initialization and HTTP server setup.
package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"code/internal/config"
	"code/internal/db"
	"code/internal/link"
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
func setupRouter(linkHandler *link.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sentrygin.New(sentrygin.Options{Repanic: false}))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
	}))

	router.TrustedPlatform = gin.PlatformCloudflare
	_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/r/:code", linkHandler.Redirect)

	api := router.Group("/api")

	api.POST("/links", linkHandler.CreateLink)
	api.GET("/links", linkHandler.ListLinks)
	api.GET("/links/:id", linkHandler.GetLink)
	api.PUT("/links/:id", linkHandler.UpdateLink)
	api.DELETE("/links/:id", linkHandler.DeleteLink)
	api.GET("/link_visits", linkHandler.ListVisits)

	return router
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

	queries := db.New(dbConn)
	linkRepo := link.NewRepository(queries)
	linkService := link.NewService(linkRepo, cfg.BaseURL)
	linkHandler := link.NewHandler(linkService)

	router := setupRouter(linkHandler)

	if err := router.Run(":8080"); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
