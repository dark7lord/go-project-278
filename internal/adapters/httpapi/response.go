package httpapi

import (
	"time"

	"code/internal/application"
)

// linkResponse is the HTTP representation of a shortened link.
type linkResponse struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

func (h *Handler) linkResponse(link application.LinkView) linkResponse {
	return linkResponse{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
		ShortURL:    h.baseURL + "/r/" + link.ShortName,
	}
}

// mapSlice converts every item of a slice with convert.
func mapSlice[T, U any](items []T, convert func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = convert(item)
	}

	return result
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
