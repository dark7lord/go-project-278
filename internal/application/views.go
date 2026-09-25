package application

import "time"

// LinkView is the application representation of a link returned by a use case.
type LinkView struct {
	ID          int64
	OriginalURL string
	ShortName   string
}

// VisitView is the application representation of a recorded link visit.
type VisitView struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	CreatedAt time.Time `json:"created_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Status    int32     `json:"status"`
}
