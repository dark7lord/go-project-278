package httpadapter

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// ListVisits handles listing all link visits.
func (h *Handler) ListVisits(c *gin.Context) {
	rangeParam := requestRange(c)

	if rangeParam == "" {
		visits, err := h.visitService.ListLinkVisits(c.Request.Context())
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

	visits, total, err := h.visitService.ListLinkVisitsRange(
		c.Request.Context(),
		application.ListLinkVisitsQuery{
			Start: int64(start),
			End:   int64(end),
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))
		return
	}

	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", start, end, total))
	c.JSON(http.StatusOK, visits)
}
