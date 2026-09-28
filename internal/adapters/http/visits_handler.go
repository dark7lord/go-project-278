package httpadapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// ListVisits handles listing a page of link visits.
func (h *Handler) ListVisits(c *gin.Context) {
	sort, err := parseSortParam(c.Query("sort"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}
	if sort != nil {
		if _, ok := visitsSortableFields[sort.Field]; !ok {
			c.JSON(http.StatusBadRequest, errJSON(application.ErrSortField.Error()))
			return
		}
	}

	pageRange, fromHeader, err := requestRange(c, linkVisitsUnit)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	page, err := h.visitService.PageLinkVisits(
		c.Request.Context(),
		application.ListLinkVisitsQuery{Range: pageRange, Sort: sort},
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeRangePage(c, linkVisitsUnit, fromHeader, application.RangePage[visitResponse]{
		Items: toVisitResponses(page.Items),
		First: page.First,
		Total: page.Total,
	})
}
