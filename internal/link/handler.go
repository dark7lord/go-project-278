package link

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	var fe *FieldError
	if errors.As(err, &fe) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{fe.Field: fe.Error()}})
		return
	}
	c.JSON(http.StatusBadRequest, errJSON(err.Error()))
}

// Handler handles HTTP requests for links.
type Handler struct {
	service *Service
}

// NewHandler creates a new Handler.
func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
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

	link, err := h.service.CreateLink(c.Request.Context(), req.OriginalURL, req.ShortName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
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

	link, err := h.service.GetLinkByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
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
		links, err := h.service.ListLinks(c.Request.Context())
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

	links, total, err := h.service.ListLinksRange(c.Request.Context(), int64(start), int64(end))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))
		return
	}

	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", start, end, total))
	c.JSON(http.StatusPartialContent, links)
}

// ListVisits handles listing all link visits.
func (h *Handler) ListVisits(c *gin.Context) {
	rangeParam := requestRange(c)

	if rangeParam == "" {
		visits, err := h.service.ListLinkVisits(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, errJSON(errInternal))
			return
		}

		c.JSON(http.StatusOK, visits)

		return
	}

	start, end, err := parseRangeParam(rangeParam)
	if err != nil {
		status, msg := rangeStatus(err)
		c.JSON(status, errJSON(msg))
		return
	}

	visits, total, err := h.service.ListLinkVisitsRange(c.Request.Context(), int64(start), int64(end))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))
		return
	}

	c.Header("Content-Range", fmt.Sprintf("visits %d-%d/%d", start, end, total))
	c.JSON(http.StatusPartialContent, visits)
}

// UpdateLinkRequest represents a request to update a link.
type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name" binding:"min=3,max=32"`
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

	if _, err := h.service.UpdateLink(c.Request.Context(), id, req.OriginalURL, req.ShortName); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}
		writeFieldErrors(c, err)

		return
	}

	c.Status(http.StatusOK)
}

// DeleteLink handles link deletion.
func (h *Handler) DeleteLink(c *gin.Context) {
	paramID := c.Param("id")
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(errInvalidID))
		return
	}

	link, err := h.service.DeleteLink(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.JSON(http.StatusOK, link)
}

// Redirect handles redirecting a short name to its original URL.
func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, errJSON("invalid code"))

		return
	}

	ctx := c.Request.Context()
	link, err := h.service.GetLinkByShortName(ctx, code)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}

		c.JSON(http.StatusInternalServerError, errJSON(err.Error()))

		return
	}

	originalURL, err := normalizeURL(link.OriginalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.Redirect(http.StatusFound, originalURL)

	var referer *string
	if ref := c.Request.Referer(); ref != "" {
		referer = &ref
	}

	if _, err := h.service.CreateLinkVisit(
		c.Request.Context(),
		link.ID, c.ClientIP(),
		c.Request.UserAgent(),
		referer,
		int32(c.Writer.Status()),
	); err != nil {
		log.Printf("failed to record visit for link %d: %v", link.ID, err)
	}
}
