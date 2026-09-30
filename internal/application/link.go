package application

// Link is a shortened link: its id, the URL it leads to and its short name.
type Link struct {
	ID          int64
	OriginalURL string
	ShortName   string
}

// LinkInput carries the fields a client sets on a link, on create and update.
type LinkInput struct {
	OriginalURL string
	ShortName   string
}
