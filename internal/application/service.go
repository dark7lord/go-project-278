package application

import (
	domainlinks "code/internal/domain/links"
)

// Service implements link application use cases.
type Service struct {
	linkReader         LinkReader
	linkWriter         LinkWriter
	visitReader        VisitReader
	visitRecorder      VisitRecorder
	baseURL            string
	shortCodeGenerator ShortCodeGenerator
}

// Compile-time checks that the service implements the application use-case ports.
var _ LinkUseCase = (*Service)(nil)
var _ VisitUseCase = (*Service)(nil)
var _ UseCase = (*Service)(nil)

// NewServiceWithGenerator creates a link application service with an injected short-code generator.
func NewServiceWithGenerator(deps ServiceDeps, baseURL string, generator ShortCodeGenerator) *Service {
	if generator == nil {
		panic("application: NewServiceWithGenerator called with nil generator")
	}

	return &Service{
		linkReader:         deps.LinkReader,
		linkWriter:         deps.LinkWriter,
		visitReader:        deps.VisitReader,
		visitRecorder:      deps.VisitRecorder,
		baseURL:            baseURL,
		shortCodeGenerator: generator,
	}
}

const (
	fieldShortName   = "short_name"
	fieldOriginalURL = "original_url"
)

func normalizeURL(raw string) (string, error) {
	parsed, err := domainlinks.NewURL(raw)
	if err != nil {
		return "", err
	}

	return parsed.String(), nil
}

func (s *Service) generateShortCode() string {
	return s.shortCodeGenerator.Generate()
}

// withShortURL injects the current short URL into a link view. short_url is
// derived from the current base URL rather than persisted, so changing the
// domain never desynchronizes existing records.
func (s *Service) withShortURL(link LinkView) LinkView {
	link.ShortURL = s.baseURL + "/r/" + link.ShortName

	return link
}
