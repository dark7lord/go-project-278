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
	ID        int64
	LinkID    int64
	CreatedAt time.Time
	IP        string
	UserAgent string
	Status    int32
}
