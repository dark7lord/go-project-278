package main

import (
	"net/http"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"code/links"
)

func setupRouter(linkHandler *links.Handler) *gin.Engine {
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

	return router
}
