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
	shortCodeGenerator ShortCodeGenerator
}

// Compile-time checks that the service implements the application use-case ports.
var _ LinkUseCase = (*Service)(nil)
var _ VisitUseCase = (*Service)(nil)
var _ UseCase = (*Service)(nil)

// NewServiceWithGenerator creates a link application service with an injected short-code generator.
func NewServiceWithGenerator(deps ServiceDeps, generator ShortCodeGenerator) *Service {
	if generator == nil {
		panic("application: NewServiceWithGenerator called with nil generator")
	}

	return &Service{
		linkReader:         deps.LinkReader,
		linkWriter:         deps.LinkWriter,
		visitReader:        deps.VisitReader,
		visitRecorder:      deps.VisitRecorder,
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
