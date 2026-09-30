package application

// Service implements link application use cases.
type Service struct {
	links    LinkStore
	visits   VisitStore
	generate func() string
}

// Compile-time check that the service implements the use-case port.
var _ UseCase = (*Service)(nil)

// NewService creates the link application service; generate proposes short
// names for links created or updated without one.
func NewService(links LinkStore, visits VisitStore, generate func() string) *Service {
	return &Service{links: links, visits: visits, generate: generate}
}

const (
	fieldShortName   = "short_name"
	fieldOriginalURL = "original_url"
)
