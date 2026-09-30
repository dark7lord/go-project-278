package application

import (
	"context"
	"errors"
	"fmt"

	domainlinks "code/internal/domain/links"
)

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

// CreateLink creates a new link with the given URL and optional short name.
func (s *Service) CreateLink(ctx context.Context, in LinkInput) (Link, error) {
	return s.saveLinkFields(ctx, in, s.links.CreateLink)
}

// Redirect resolves a short name and records the visit.
func (s *Service) Redirect(ctx context.Context, shortName string, visit VisitInput) (Link, error) {
	link, err := s.GetLinkByShortName(ctx, shortName)
	if err != nil {
		return Link{}, err
	}

	if _, err := s.visits.CreateLinkVisit(ctx, link.ID, visit); err != nil {
		return Link{}, fmt.Errorf("record link visit: %w", err)
	}

	return link, nil
}

// GetLinkByID retrieves a link by its ID.
func (s *Service) GetLinkByID(ctx context.Context, id int64) (Link, error) {
	link, err := s.links.GetLinkByID(ctx, id)
	if err != nil {
		return Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// GetLinkByShortName retrieves a link by its short name.
func (s *Service) GetLinkByShortName(ctx context.Context, shortName string) (Link, error) {
	link, err := s.links.GetLinkByShortName(ctx, shortName)
	if err != nil {
		return Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// PageLinks retrieves a paginated page of links.
func (s *Service) PageLinks(ctx context.Context, q PageQuery) (RangePage[Link], error) {
	return s.links.PageLinks(ctx, q)
}

// PageLinkVisits retrieves a paginated page of link visits.
func (s *Service) PageLinkVisits(ctx context.Context, q PageQuery) (RangePage[Visit], error) {
	return s.visits.PageLinkVisits(ctx, q)
}

// UpdateLink updates an existing link.
func (s *Service) UpdateLink(ctx context.Context, id int64, in LinkInput) (Link, error) {
	persist := func(ctx context.Context, normalizedURL, name string) (Link, error) {
		return s.links.UpdateLink(ctx, id, normalizedURL, name)
	}

	return s.saveLinkFields(ctx, in, persist)
}

// DeleteLink deletes a link by its ID.
func (s *Service) DeleteLink(ctx context.Context, id int64) (Link, error) {
	link, err := s.links.DeleteLink(ctx, id)
	if err != nil {
		return Link{}, fmt.Errorf("delete link: %w", err)
	}

	return link, nil
}

// saveLinkFields validates the fields shared by create and update and persists
// them through persist, which differs only in the writer call per operation.
func (s *Service) saveLinkFields(
	ctx context.Context,
	in LinkInput,
	persist func(ctx context.Context, normalizedURL, name string) (Link, error),
) (Link, error) {
	normalized, err := domainlinks.NormalizeURL(in.OriginalURL)
	if err != nil {
		return Link{}, &FieldError{
			Field: fieldOriginalURL,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidURL, in.OriginalURL),
		}
	}

	if in.ShortName == "" {
		return s.withGeneratedShortName(ctx, func(ctx context.Context, name string) (Link, error) {
			return persist(ctx, normalized, name)
		})
	}

	code, err := domainlinks.NormalizeShortCode(in.ShortName)
	if err != nil {
		return Link{}, &FieldError{
			Field: fieldShortName,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidShortCode, in.ShortName),
		}
	}

	link, err := persist(ctx, normalized, code)
	if err != nil {
		return Link{}, fmt.Errorf("persist link: %w", err)
	}

	return link, nil
}

// maxShortCodeAttempts bounds retries when a generated short name collides.
const maxShortCodeAttempts = 10

// withGeneratedShortName keeps proposing generated short names until the
// storage accepts one or the attempt budget is exhausted. Only short-name
// collisions are retried; any other error ends the loop immediately.
func (s *Service) withGeneratedShortName(
	ctx context.Context,
	try func(ctx context.Context, shortName string) (Link, error),
) (Link, error) {
	for attempt := 0; attempt < maxShortCodeAttempts; attempt++ {
		link, err := try(ctx, s.generate())
		if err == nil {
			return link, nil
		}

		var fieldErr *FieldError
		if !errors.As(err, &fieldErr) || !errors.Is(fieldErr.Err, ErrShortNameAlreadyUse) {
			return Link{}, fmt.Errorf("persist link: %w", err)
		}
	}

	return Link{}, fmt.Errorf(
		"%w: exhausted %d attempts",
		ErrShortCodeGenerationFailed,
		maxShortCodeAttempts,
	)
}
