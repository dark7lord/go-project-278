// Package httpapi provides the Gin HTTP transport for link use cases.
package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// Handler handles HTTP requests for links and their visits.
type Handler struct {
	service application.UseCase
	baseURL string
}

// NewHandler creates a Handler from the use cases and the public link origin.
func NewHandler(service application.UseCase, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

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

// mapSlice converts every item of a slice with convert.
func mapSlice[T, U any](items []T, convert func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = convert(item)
	}

	return result
}
