package main

import (
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"code/links"
)

func setupRouter(linkHandler *links.Handler, frontendURL string) *gin.Engine {
	router := gin.Default()
	router.Use(sentrygin.New(sentrygin.Options{Repanic: false}))
	origins := []string{"http://localhost:5173"}
	if frontendURL != "" {
		origins = append(origins, frontendURL)
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
