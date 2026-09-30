// Package httpapi provides the Gin HTTP transport for link use cases.
package httpapi

import (
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
