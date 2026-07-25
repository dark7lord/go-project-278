// Package main implements a service that processes user requests.
package main

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{Repanic: false}))

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
		panic("I love Haruno Sakura")
	})

	return router
}

func main() {
	r := setupRouter()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		log.Fatal("SENTRY_DSN is not set")
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	}); err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	r.GET("/debug-sentry", func(c *gin.Context) {
		panic("test sentry capture")
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
