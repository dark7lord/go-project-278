package application

import (
	"context"
	"errors"
	"fmt"

	"code/internal/domain/links"
)

// Service implements UseCase over the stores.
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

// CreateLink stores a new link, generating a short name when in has none.
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

// GetLinkByID reads one link.
func (s *Service) GetLinkByID(ctx context.Context, id int64) (Link, error) {
	link, err := s.links.GetLinkByID(ctx, id)
	if err != nil {
		return Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// GetLinkByShortName reads one link.
func (s *Service) GetLinkByShortName(ctx context.Context, shortName string) (Link, error) {
	link, err := s.links.GetLinkByShortName(ctx, shortName)
	if err != nil {
		return Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// PageLinks reads a page of links.
func (s *Service) PageLinks(ctx context.Context, q PageQuery) (RangePage[Link], error) {
	return s.links.PageLinks(ctx, q)
}

// PageLinkVisits reads a page of visits.
func (s *Service) PageLinkVisits(ctx context.Context, q PageQuery) (RangePage[Visit], error) {
	return s.visits.PageLinkVisits(ctx, q)
}

// UpdateLink rewrites link id, checking in as CreateLink does.
func (s *Service) UpdateLink(ctx context.Context, id int64, in LinkInput) (Link, error) {
	persist := func(ctx context.Context, normalizedURL, name string) (Link, error) {
		return s.links.UpdateLink(ctx, id, normalizedURL, name)
	}

	return s.saveLinkFields(ctx, in, persist)
}

// DeleteLink removes link id.
func (s *Service) DeleteLink(ctx context.Context, id int64) (Link, error) {
	link, err := s.links.DeleteLink(ctx, id)
	if err != nil {
		return Link{}, fmt.Errorf("delete link: %w", err)
	}

	return link, nil
}

// saveLinkFields validates in and stores it through persist, the one step
// create and update do differently.
func (s *Service) saveLinkFields(
	ctx context.Context,
	in LinkInput,
	persist func(ctx context.Context, normalizedURL, name string) (Link, error),
) (Link, error) {
	normalized, err := links.NormalizeURL(in.OriginalURL)
	if err != nil {
		return Link{}, &FieldError{
			Field: fieldOriginalURL,
			Err:   fmt.Errorf("%w: %s", links.ErrInvalidURL, in.OriginalURL),
		}
	}

	if in.ShortName == "" {
		return s.withGeneratedShortName(ctx, func(ctx context.Context, name string) (Link, error) {
			return persist(ctx, normalized, name)
		})
	}

	code, err := links.NormalizeShortCode(in.ShortName)
	if err != nil {
		return Link{}, &FieldError{
			Field: fieldShortName,
			Err:   fmt.Errorf("%w: %s", links.ErrInvalidShortCode, in.ShortName),
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

// withGeneratedShortName retries try with fresh names while they collide, up to
// maxShortCodeAttempts; any other error stops it at once.
func (s *Service) withGeneratedShortName(
	ctx context.Context,
	try func(ctx context.Context, shortName string) (Link, error),
) (Link, error) {
	for range maxShortCodeAttempts {
		link, err := try(ctx, s.generate())
		if err == nil {
			return link, nil
		}
		if !errors.Is(err, ErrShortNameAlreadyUse) {
			return Link{}, fmt.Errorf("persist link: %w", err)
		}
	}

	return Link{}, fmt.Errorf(
		"%w: exhausted %d attempts",
		ErrShortCodeGenerationFailed,
		maxShortCodeAttempts,
	)
}
