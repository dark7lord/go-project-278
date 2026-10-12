package application

import "time"

// Link is a shortened link: its id, the URL it leads to,
// its short name and when the database stored it.
type Link struct {
	ID          int64
	OriginalURL string
	ShortName   string
	CreatedAt   time.Time
}

// LinkInput carries the fields a client sets on a link, on create and update.
type LinkInput struct {
	OriginalURL string
	ShortName   string
}
