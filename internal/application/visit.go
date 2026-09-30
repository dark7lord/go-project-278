package application

import "time"

// Visit is a recorded follow of a link.
type Visit struct {
	ID        int64
	LinkID    int64
	CreatedAt time.Time
	IP        string
	UserAgent string
	Referer   *string
	Status    int32
}

// VisitInput describes one followed link: who came, from where, and the status
// the redirect answered with.
type VisitInput struct {
	IP        string
	UserAgent string
	Referer   *string
	Status    int32
}
