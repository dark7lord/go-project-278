// Package httpadapter provides the Gin HTTP transport for link use cases.
package httpadapter

import (
	"code/internal/application"
)

// Handler handles HTTP requests for links.
type Handler struct {
	linkService  application.LinkUseCase
	visitService application.VisitUseCase
	baseURL      string
}

// NewHandler creates a Handler from use cases and the public link origin.
func NewHandler(
	linkService application.LinkUseCase,
	visitService application.VisitUseCase,
	baseURL string,
) *Handler {
	return &Handler{
		linkService:  linkService,
		visitService: visitService,
		baseURL:      baseURL,
	}
}
