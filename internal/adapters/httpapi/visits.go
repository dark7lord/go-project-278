package httpapi

import (
	"net/http"
	"time"

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

// visitResponse is the HTTP representation of a recorded link visit.
type visitResponse struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	CreatedAt time.Time `json:"created_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	// "reffer", not "referer": the dashboard's visits column reads this name
	Reffer *string `json:"reffer"`
	Status int32   `json:"status"`
}

func toVisitResponse(visit application.VisitView) visitResponse {
	return visitResponse{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt,
		IP:        visit.IP,
		UserAgent: visit.UserAgent,
		Reffer:    visit.Referer,
		Status:    visit.Status,
	}
}
