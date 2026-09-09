// Package httpadapter provides the Gin HTTP transport for link use cases.
package httpadapter

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"code/internal/application"
)

const errInvalidID = "invalid id"
const errInternal = "internal error"

func errJSON(msg string) gin.H {
	return gin.H{"error": msg}
}

var camelRe = regexp.MustCompile(`([a-z0-9])([A-Z])`)

// toSnakeCase converts "ShortName" to "short_name" and "OriginalURL" to "original_url".
func toSnakeCase(s string) string {
	return strings.ToLower(camelRe.ReplaceAllString(s, `${1}_${2}`))
}

// bindMessage translates a validator tag into a human-readable message.
func bindMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	default:
		return fe.Error()
	}
}

// writeBindErrors returns 422 for validator errors, otherwise 400 for invalid JSON.
func writeBindErrors(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		result := make(map[string]string)
		for _, fieldErr := range validationErrors {
			result[toSnakeCase(fieldErr.Field())] = bindMessage(fieldErr)
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": result})

		return
	}
	c.JSON(http.StatusBadRequest, errJSON("invalid request"))
}

// writeFieldErrors returns 422 for field errors, otherwise 400.
func writeFieldErrors(c *gin.Context, err error) {
	var fe *application.FieldError
	if errors.As(err, &fe) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{fe.Field: fe.Error()}})
		return
	}
	c.JSON(http.StatusBadRequest, errJSON(err.Error()))
}

// Handler handles HTTP requests for links.
type Handler struct {
	linkService  application.LinkUseCase
	visitService application.VisitUseCase
}

// NewHandler creates a new Handler from separate link and visit use cases.
func NewHandler(linkService application.LinkUseCase, visitService application.VisitUseCase) *Handler {
	return &Handler{linkService: linkService, visitService: visitService}
}

// CreateLinkRequest represents a request to create a link.
type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

// CreateLink handles link creation.
func (h *Handler) CreateLink(c *gin.Context) {
	var req CreateLinkRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindErrors(c, err)
		return
	}

	cmd := application.CreateLinkCommand{OriginalURL: req.OriginalURL, ShortName: req.ShortName}
	link, err := h.linkService.CreateLink(c.Request.Context(), cmd)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}
		writeFieldErrors(c, err)

		return
	}

	c.JSON(http.StatusCreated, link)
}

// GetLink handles link retrieval by ID.
func (h *Handler) GetLink(c *gin.Context) {
	paramID := c.Param("id")
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(errInvalidID))
		return
	}

	link, err := h.linkService.GetLinkByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}

		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.JSON(http.StatusOK, link)
}

// rangeStatus maps a range parsing error to an HTTP status and message.
func rangeStatus(err error) (int, string) {
	if errors.Is(err, ErrRangeNotSatisfiable) {
		return http.StatusRequestedRangeNotSatisfiable, err.Error()
	}

	return http.StatusBadRequest, err.Error()
}

// requestRange returns the range parameter from query string or Range header.
func requestRange(c *gin.Context) string {
	if q := c.Query("range"); q != "" {
		return q
	}

	return c.GetHeader("Range")
}

// ListLinks handles listing all links.
func (h *Handler) ListLinks(c *gin.Context) {
	rangeParam := requestRange(c)

	if rangeParam == "" {
		links, err := h.linkService.ListLinks(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, errJSON("internal server error"))
			return
		}

		c.JSON(http.StatusOK, links)

		return
	}

	start, end, err := parseRangeParam(rangeParam)
	if err != nil {
		status, msg := rangeStatus(err)
		c.JSON(status, errJSON(msg))
		return
	}

	links, total, err := h.linkService.ListLinksRange(c.Request.Context(), application.ListLinksQuery{Start: int64(start), End: int64(end)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))
		return
	}

	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", start, end, total))
	c.JSON(http.StatusOK, links)
}

// UpdateLinkRequest represents a request to update a link.
type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

// UpdateLink handles link updates.
func (h *Handler) UpdateLink(c *gin.Context) {
	paramID := c.Param("id")
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(errInvalidID))
		return
	}

	var req UpdateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindErrors(c, err)
		return
	}

	cmd := application.UpdateLinkCommand{OriginalURL: req.OriginalURL, ShortName: req.ShortName}
	updated, err := h.linkService.UpdateLink(c.Request.Context(), id, cmd)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}
		writeFieldErrors(c, err)

		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteLink handles link deletion.
func (h *Handler) DeleteLink(c *gin.Context) {
	paramID := c.Param("id")
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(errInvalidID))
		return
	}

	_, err = h.linkService.DeleteLink(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.Status(http.StatusNoContent)
}

// Redirect handles redirecting a short name to its original URL.
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")

	ctx := c.Request.Context()
	var referer *string
	if ref := c.Request.Referer(); ref != "" {
		referer = &ref
	}

	link, err := h.linkService.Redirect(ctx, application.RedirectCommand{
		ShortName: code,
		VisitMeta: application.VisitMeta{
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Referer:   referer,
		},
	})

	if err != nil {
		if errors.Is(err, application.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}

		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL)
}
