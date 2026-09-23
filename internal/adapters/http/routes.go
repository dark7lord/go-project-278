package httpadapter

import (
	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers the JSON API routes on the given router group.
func (h *Handler) RegisterAPIRoutes(api *gin.RouterGroup) {
	api.POST("/links", h.CreateLink)
	api.GET("/links", h.ListLinks)
	api.GET("/links/:id", h.GetLink)
	api.PUT("/links/:id", h.UpdateLink)
	api.DELETE("/links/:id", h.DeleteLink)
	api.GET("/link_visits", h.ListVisits)
}
