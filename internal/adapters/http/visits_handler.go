package httpadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// ListVisits handles listing all link visits.
func (h *Handler) ListVisits(c *gin.Context) {
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
		if _, ok := visitsSortableFields[sort.Field]; !ok {
			c.JSON(http.StatusBadRequest, errJSON(application.ErrSortField.Error()))
			return
		}
	}

	if rangeParam == "" {
		visits, err := h.visitService.ListLinkVisits(c.Request.Context())
		if err != nil {
			writeServiceError(c, err)
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

	page, err := h.visitService.PageLinkVisits(
		c.Request.Context(),
		application.ListLinkVisitsQuery{
			Start: start,
			End:   end,
			Sort:  sort,
		},
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeRangePage(c, "link_visits", page)
}
