// Package httpadapter provides the Gin HTTP transport for link use cases.
package httpadapter

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"code/internal/application"
)

const errInvalidID = "invalid id"
const errInternal = "internal error"

func errJSON(msg string) gin.H {
	return gin.H{"error": msg}
}

// parsePositiveID parses an id path parameter into a positive int64.
func parsePositiveID(paramID string) (int64, error) {
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New(errInvalidID)
	}

	return id, nil
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

// writeServiceError maps an application error to a stable HTTP response.
// Only classified errors are exposed to the client; everything else is logged
// and reported as a generic 500 so internal details never leak.
func writeServiceError(c *gin.Context, err error) {
	var fe *application.FieldError
	if errors.As(err, &fe) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{fe.Field: fe.Error()}})
		return
	}

	if errors.Is(err, application.ErrSortField) {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	if errors.Is(err, application.ErrNotFound) {
		c.JSON(http.StatusNotFound, errJSON(err.Error()))
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		c.JSON(http.StatusServiceUnavailable, errJSON("request timeout"))
		return
	}

	_ = c.Error(err)
	sentry.CaptureException(err)
	c.JSON(http.StatusInternalServerError, errJSON(errInternal))
}

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

// linkResponse is the HTTP representation of a shortened link.
type linkResponse struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

func (h *Handler) linkResponse(link application.LinkView) linkResponse {
	return linkResponse{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
		ShortURL:    h.baseURL + "/r/" + link.ShortName,
	}
}

func (h *Handler) linkResponses(links []application.LinkView) []linkResponse {
	responses := make([]linkResponse, len(links))
	for index, link := range links {
		responses[index] = h.linkResponse(link)
	}

	return responses
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
		writeServiceError(c, err)

		return
	}

	c.JSON(http.StatusCreated, h.linkResponse(link))
}

// GetLink handles link retrieval by ID.
func (h *Handler) GetLink(c *gin.Context) {
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	link, err := h.linkService.GetLinkByID(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.JSON(http.StatusOK, h.linkResponse(link))
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

	sort, err := parseSortParam(c.Query("sort"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}
	if sort != nil {
		if rangeParam == "" {
			c.JSON(http.StatusBadRequest, errJSON(ErrSortWithoutRange.Error()))
			return
		}
		if _, ok := linksSortableFields[sort.Field]; !ok {
			c.JSON(http.StatusBadRequest, errJSON(application.ErrSortField.Error()))
			return
		}
	}

	if rangeParam == "" {
		links, err := h.linkService.ListLinks(c.Request.Context())
		if err != nil {
			writeServiceError(c, err)
			return
		}

		c.JSON(http.StatusOK, h.linkResponses(links))

		return
	}

	start, end, err := parseRangeParam(rangeParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	page, err := h.linkService.ListLinksRange(
		c.Request.Context(),
		application.ListLinksQuery{Start: start, End: end, Sort: sort},
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeRangePage(c, "links", application.RangePage[linkResponse]{
		Items: h.linkResponses(page.Items),
		Start: page.Start,
		Total: page.Total,
	})
}

// UpdateLinkRequest represents a request to update a link.
type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

// UpdateLink handles link updates.
func (h *Handler) UpdateLink(c *gin.Context) {
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
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
		writeServiceError(c, err)

		return
	}

	c.JSON(http.StatusOK, h.linkResponse(updated))
}

// DeleteLink handles link deletion.
func (h *Handler) DeleteLink(c *gin.Context) {
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	_, err = h.linkService.DeleteLink(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)

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
		writeServiceError(c, err)

		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL)
}
