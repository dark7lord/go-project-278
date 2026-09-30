package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// redirectStatus is the single source for the response status and for the
// status recorded on the visit, so the two can never drift apart.
const redirectStatus = http.StatusFound

// CreateLink handles link creation.
func (h *Handler) CreateLink(c *gin.Context) {
	req, ok := bindLinkRequest(c)
	if !ok {
		return
	}

	in := application.LinkInput{OriginalURL: req.OriginalURL, ShortName: req.ShortName}
	link, err := h.service.CreateLink(c.Request.Context(), in)
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

	link, err := h.service.GetLinkByID(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.JSON(http.StatusOK, h.linkResponse(link))
}

// ListLinks handles listing a page of links.
func (h *Handler) ListLinks(c *gin.Context) {
	sort, err := parseSortParam(c.Query("sort"), linksSortFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	pageRange, fromHeader, err := requestRange(c, linksUnit)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	page, err := h.service.PageLinks(
		c.Request.Context(),
		application.PageQuery{Range: pageRange, Sort: sort},
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeRangePage(c, linksUnit, fromHeader, application.RangePage[linkResponse]{
		Items: mapSlice(page.Items, h.linkResponse),
		First: page.First,
		Total: page.Total,
	})
}

// UpdateLink handles link updates.
func (h *Handler) UpdateLink(c *gin.Context) {
	id, err := parsePositiveID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	req, ok := bindLinkRequest(c)
	if !ok {
		return
	}

	in := application.LinkInput{OriginalURL: req.OriginalURL, ShortName: req.ShortName}
	updated, err := h.service.UpdateLink(c.Request.Context(), id, in)
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

	_, err = h.service.DeleteLink(c.Request.Context(), id)
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.Status(http.StatusNoContent)
}

// Redirect handles redirecting a short name to its original URL.
func (h *Handler) Redirect(c *gin.Context) {
	var referer *string
	if ref := c.Request.Referer(); ref != "" {
		referer = &ref
	}

	link, err := h.service.Redirect(c.Request.Context(), c.Param("code"), application.VisitInput{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   referer,
		Status:    int32(redirectStatus),
	})
	if err != nil {
		writeServiceError(c, err)

		return
	}

	c.Redirect(redirectStatus, link.OriginalURL)
}

const errInvalidID = "invalid id"

// parsePositiveID parses an id path parameter into a positive int64.
func parsePositiveID(paramID string) (int64, error) {
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New(errInvalidID)
	}

	return id, nil
}

// linkRequest represents the request body shared by link creation and update.
type linkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

// bindLinkRequest decodes the link body shared by create and update, writing
// the error response itself when the body is invalid.
func bindLinkRequest(c *gin.Context) (linkRequest, bool) {
	var req linkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindErrors(c, err)

		return req, false
	}

	return req, true
}

// linkResponse is the HTTP representation of a shortened link.
type linkResponse struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

func (h *Handler) linkResponse(link application.Link) linkResponse {
	return linkResponse{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
		ShortURL:    h.baseURL + "/r/" + link.ShortName,
	}
}
