package httpadapter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// visitResponse is the HTTP representation of a recorded link visit.
type visitResponse struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	CreatedAt time.Time `json:"created_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Status    int32     `json:"status"`
}

func toVisitResponse(visit application.VisitView) visitResponse {
	return visitResponse{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt,
		IP:        visit.IP,
		UserAgent: visit.UserAgent,
		Status:    visit.Status,
	}
}

func toVisitResponses(visits []application.VisitView) []visitResponse {
	responses := make([]visitResponse, len(visits))
	for index, visit := range visits {
		responses[index] = toVisitResponse(visit)
	}

	return responses
}

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

		c.JSON(http.StatusOK, toVisitResponses(visits))

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

	writeRangePage(c, "link_visits", application.RangePage[visitResponse]{
		Items: toVisitResponses(page.Items),
		Start: page.Start,
		Total: page.Total,
	})
}
