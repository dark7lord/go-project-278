package httpadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRootRoutes registers the non-API routes on the given router.
func (h *Handler) RegisterRootRoutes(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	router.GET("/r/:code", h.Redirect)
}

// RegisterAPIRoutes registers the JSON API routes on the given router group.
func (h *Handler) RegisterAPIRoutes(api *gin.RouterGroup) {
	api.POST("/links", h.CreateLink)
	api.GET("/links", h.ListLinks)
	api.GET("/links/:id", h.GetLink)
	api.PUT("/links/:id", h.UpdateLink)
	api.DELETE("/links/:id", h.DeleteLink)
	api.GET("/link_visits", h.ListVisits)
}
