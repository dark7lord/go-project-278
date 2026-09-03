package links

// Visit records a request made through a shortened link.
type Visit struct {
	ID        int64
	LinkID    int64
	IP        string
	UserAgent string
	Referer   *string
	Status    int32
}
