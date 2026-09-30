package application

// LinkInput carries the fields a client sets on a link, on create and update.
type LinkInput struct {
	OriginalURL string
	ShortName   string
}

// Visit describes one followed link: who came, from where, and the status the
// redirect answered with.
type Visit struct {
	IP        string
	UserAgent string
	Referer   *string
	Status    int32
}
