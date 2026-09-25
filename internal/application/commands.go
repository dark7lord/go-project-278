package application

// CreateLinkCommand represents the intent to create a new shortened link.
type CreateLinkCommand struct {
	OriginalURL string
	ShortName   string
}

// UpdateLinkCommand represents the intent to update an existing shortened link.
type UpdateLinkCommand struct {
	OriginalURL string
	ShortName   string
}

// RedirectCommand describes a redirect request, its visit metadata and the
// response status the visit is recorded with.
type RedirectCommand struct {
	ShortName string
	VisitMeta VisitMeta
	Status    int32
}
