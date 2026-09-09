package application

import (
	"fmt"
	"math/rand"

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

type defaultShortCodeGenerator struct{}

func (defaultShortCodeGenerator) Generate() string {
	return genRandomName()
}

// NewService creates a link application service.
func NewService(deps ServiceDeps, baseURL string) *Service {
	return NewServiceWithGenerator(deps, baseURL, defaultShortCodeGenerator{})
}

// NewServiceWithGenerator creates a link application service with an injected short-code generator.
func NewServiceWithGenerator(deps ServiceDeps, baseURL string, generator ShortCodeGenerator) *Service {
	if generator == nil {
		generator = defaultShortCodeGenerator{}
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

var colors = []string{"red", "orange", "yellow", "green", "cyan", "blue", "purple"}

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

func genRandomName() string {
	prefixColor := colors[rand.Intn(len(colors))]
	infixColor := colors[rand.Intn(len(colors))]
	suffixNum := rand.Intn(1024)

	return fmt.Sprintf("%s-%s-link-%d", prefixColor, infixColor, suffixNum)
}

func (s *Service) generateShortLink(shortName string) string {
	return s.baseURL + "/r/" + shortName
}

func (s *Service) generateShortCode() string {
	return s.shortCodeGenerator.Generate()
}
