package links

type Visit struct {
	ID        int64
	LinkID    int64
	IP        string
	UserAgent string
	Referer   *string
	Status    int32
}
