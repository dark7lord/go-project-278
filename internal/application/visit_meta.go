package application

// VisitMeta contains metadata captured at the redirect boundary.
type VisitMeta struct {
	IP        string
	UserAgent string
	Referer   *string
}
