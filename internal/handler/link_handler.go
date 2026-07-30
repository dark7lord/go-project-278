// Package handler provides HTTP handlers for link management.
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"

	"code/internal/service"
)

const errInvalidID = "invalid id"
const errInternal = "internal error"

func errJSON(msg string) gin.H {
	return gin.H{"error": msg}
}

// LinkHandler handles HTTP requests for links.
type LinkHandler struct {
	service *service.LinkService
}

// NewLinkHandler creates a new LinkHandler.
func NewLinkHandler(s *service.LinkService) *LinkHandler {
	return &LinkHandler{service: s}
}

// CreateLinkRequest represents a request to create a link.
type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name"`
}

// CreateLink handles link creation.
func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	link, err := h.service.CreateLink(c.Request.Context(), req.OriginalURL, req.ShortName)
	if err != nil {
		if errors.Is(err, service.ErrLinkExists) {
			c.JSON(http.StatusConflict, errJSON(err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))

		return
	}

	c.JSON(http.StatusCreated, link)
}

// GetLink handles link retrieval by ID.
func (h *LinkHandler) GetLink(c *gin.Context) {
	paramID := c.Param("id")
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(errInvalidID))
		return
	}

	link, err := h.service.GetLink(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}

		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.JSON(http.StatusOK, link)
}

var rangeRe = regexp.MustCompile(`\[(\d+),(\d+)\]`) // captures: [full, start, end]

// ListLinks handles listing all links.
func (h *LinkHandler) ListLinks(c *gin.Context) {
	rangeParam := c.Query("range")

	if rangeParam == "" {
		links, err := h.service.ListLinks(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusOK, links)

		return
	}

	matches := rangeRe.FindStringSubmatch(rangeParam)

	if len(matches) != 3 {
		c.JSON(http.StatusBadRequest, errJSON("invalid range, expected [start,end]"))
		return
	}

	start, err := strconv.Atoi(matches[1])
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON("invalid start value"))
		return
	}
	end, err := strconv.Atoi(matches[2])
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON("invalid end value"))
		return
	}

	if start < 0 || start > end {
		c.JSON(http.StatusRequestedRangeNotSatisfiable, errJSON("range not satisfiable"))
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

// UpdateLinkRequest represents a request to update a link.
type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

// UpdateLink handles link updates.
func (h *LinkHandler) UpdateLink(c *gin.Context) {
	paramID := c.Param("id")
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(errInvalidID))
		return
	}

	var req UpdateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	if _, err := h.service.UpdateLink(c.Request.Context(), id, req.OriginalURL, req.ShortName); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}
		if errors.Is(err, service.ErrLinkExists) {
			c.JSON(http.StatusConflict, errJSON(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.Status(http.StatusOK)
}

// DeleteLink handles link deletion.
func (h *LinkHandler) DeleteLink(c *gin.Context) {
	paramID := c.Param("id")
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(errInvalidID))
		return
	}

	if err := h.service.DeleteLink(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, errJSON(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))

		return
	}

	c.Status(http.StatusNoContent)
}
