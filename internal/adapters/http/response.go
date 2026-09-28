package httpadapter

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

func (h *Handler) linkResponses(links []application.LinkView) []linkResponse {
	responses := make([]linkResponse, len(links))
	for index, link := range links {
		responses[index] = h.linkResponse(link)
	}

	return responses
}

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
