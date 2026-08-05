// Package main is the entry point for the URL shortener service.
package main

import (
	"log"

	"github.com/joho/godotenv"

	"code/internal/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
