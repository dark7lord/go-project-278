package httpadapter

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

const errInvalidID = "invalid id"

// parsePositiveID parses an id path parameter into a positive int64.
func parsePositiveID(paramID string) (int64, error) {
	id, err := strconv.ParseInt(paramID, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New(errInvalidID)
	}

	return id, nil
}

// LinkRequest represents the request body shared by link creation and update.
type LinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

// bindLinkRequest decodes the link body shared by create and update, writing
// the error response itself when the body is invalid.
func bindLinkRequest(c *gin.Context) (LinkRequest, bool) {
	var req LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBindErrors(c, err)

		return req, false
	}

	return req, true
}

// requestRange returns the range parameter from query string or Range header.
func requestRange(c *gin.Context) string {
	if q := c.Query("range"); q != "" {
		return q
	}

	return c.GetHeader("Range")
}
