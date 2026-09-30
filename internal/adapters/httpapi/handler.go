// Package httpapi serves the use cases as a JSON API over gin.
package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// Handler serves the links and visits API.
type Handler struct {
	service application.UseCase
	baseURL string
}

// NewHandler serves service; short URLs start with baseURL, the public origin.
func NewHandler(service application.UseCase, baseURL string) *Handler {
	return &Handler{service: service, baseURL: baseURL}
}

// RegisterRootRoutes adds /ping and the /r/:code redirect.
func (h *Handler) RegisterRootRoutes(router *gin.Engine) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	router.GET("/r/:code", h.Redirect)
}

// RegisterAPIRoutes adds the JSON API to api, the /api group.
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
