// Package main implements a service that processes user requests.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"

	"code/db"
	"code/links"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg, err := loadConfig()
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
	linkRepo := links.NewRepository(queries)
	linkService := links.NewService(linkRepo, cfg.BaseURL)
	linkHandler := links.NewHandler(linkService)

	router := setupRouter(linkHandler)

	if err := router.Run(":8080"); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
