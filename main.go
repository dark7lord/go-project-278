// Package main implements a service that processes user requests.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"code/internal/db"
	"code/internal/handler"
	"code/internal/repository"
	"code/internal/service"
)

func setupRouter(linkHandler *handler.LinkHandler) *gin.Engine {
	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{Repanic: false}))
	origins := []string{"http://localhost:5173"}
	if fe := os.Getenv("FRONTEND_URL"); fe != "" {
		origins = append(origins, fe)
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins: origins,
	}))

	api := router.Group("/api")

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	api.POST("/links", linkHandler.CreateLink)
	api.GET("/links", linkHandler.ListLinks)
	api.GET("/links/:id", linkHandler.GetLink)
	api.PUT("/links/:id", linkHandler.UpdateLink)
	api.DELETE("/links/:id", linkHandler.DeleteLink)

	return router
}

func connectDB(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return conn, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	sentryDsn := os.Getenv("SENTRY_DSN")
	if sentryDsn == "" {
		log.Fatal("SENTRY_DSN is not set")
	}

	if err := sentry.Init(sentry.ClientOptions{Dsn: sentryDsn}); err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	dbDsn := os.Getenv("DATABASE_URL")
	if dbDsn == "" {
		log.Fatal("DATABASE_URL is not set") //nolint:gocritic
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("BASE_URL is not set")
	}

	dbConn, err := connectDB(dbDsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbConn.Close() //nolint:errcheck

	queries := db.New(dbConn)
	linkRepo := repository.NewLinkRepository(queries)
	linkService := service.NewLinkService(linkRepo, baseURL)
	linkHandler := handler.NewLinkHandler(linkService)

	router := setupRouter(linkHandler)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
