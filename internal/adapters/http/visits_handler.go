package httpadapter

import (
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
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	page, err := h.visitService.ListLinkVisitsRange(
		c.Request.Context(),
		application.ListLinkVisitsQuery{
			Start: start,
			End:   end,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errJSON(errInternal))
		return
	}

	writeRangePage(c, "link_visits", page)
}
