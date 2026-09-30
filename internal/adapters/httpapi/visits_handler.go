package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// ListVisits handles listing a page of link visits.
func (h *Handler) ListVisits(c *gin.Context) {
	sort, err := parseSortParam(c.Query("sort"), visitsSortFields)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	pageRange, fromHeader, err := requestRange(c, linkVisitsUnit)
	if err != nil {
		c.JSON(http.StatusBadRequest, errJSON(err.Error()))
		return
	}

	page, err := h.service.PageLinkVisits(
		c.Request.Context(),
		application.PageQuery{Range: pageRange, Sort: sort},
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeRangePage(c, linkVisitsUnit, fromHeader, application.RangePage[visitResponse]{
		Items: mapSlice(page.Items, toVisitResponse),
		First: page.First,
		Total: page.Total,
	})
}
