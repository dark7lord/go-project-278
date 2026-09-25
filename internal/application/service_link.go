package application

import (
	"context"
	"errors"
	"fmt"

	domainlinks "code/internal/domain/links"
)

// CreateLink creates a new link with the given URL and optional short name.
func (s *Service) CreateLink(ctx context.Context, cmd CreateLinkCommand) (LinkView, error) {
	return s.saveLinkFields(ctx, cmd.OriginalURL, cmd.ShortName, s.linkWriter.CreateLink)
}

// Redirect resolves a short name, validates its destination, and records the visit.
func (s *Service) Redirect(ctx context.Context, cmd RedirectCommand) (LinkView, error) {
	link, err := s.GetLinkByShortName(ctx, cmd.ShortName)
	if err != nil {
		return LinkView{}, err
	}

	if _, err := domainlinks.NewURL(link.OriginalURL); err != nil {
		return LinkView{}, fmt.Errorf("normalize redirect URL: %w", err)
	}

	if _, err := s.createLinkVisit(
		ctx,
		link.ID,
		cmd.VisitMeta.IP,
		cmd.VisitMeta.UserAgent,
		cmd.VisitMeta.Referer,
		302,
	); err != nil {
		return LinkView{}, fmt.Errorf("record link visit: %w", err)
	}

	return link, nil
}

// GetLinkByID retrieves a link by its ID.
func (s *Service) GetLinkByID(ctx context.Context, id int64) (LinkView, error) {
	link, err := s.linkReader.GetLinkByID(ctx, id)
	if err != nil {
		return LinkView{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// GetLinkByShortName retrieves a link by its short name.
func (s *Service) GetLinkByShortName(ctx context.Context, shortName string) (LinkView, error) {
	link, err := s.linkReader.GetLinkByShortName(ctx, shortName)
	if err != nil {
		return LinkView{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// ListLinks retrieves all links.
func (s *Service) ListLinks(ctx context.Context) ([]LinkView, error) {
	links, err := s.linkReader.ListLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}

	return links, nil
}

// PageLinks retrieves a paginated page of links.
func (s *Service) PageLinks(ctx context.Context, q ListLinksQuery) (RangePage[LinkView], error) {
	page, err := s.linkReader.PageLinks(ctx, q)
	if err != nil {
		return RangePage[LinkView]{}, err
	}

	return page, nil
}

// UpdateLink updates an existing link.
func (s *Service) UpdateLink(ctx context.Context, id int64, cmd UpdateLinkCommand) (LinkView, error) {
	persist := func(ctx context.Context, normalizedURL, name string) (LinkView, error) {
		return s.linkWriter.UpdateLink(ctx, id, normalizedURL, name)
	}

	return s.saveLinkFields(ctx, cmd.OriginalURL, cmd.ShortName, persist)
}

// DeleteLink deletes a link by its ID.
func (s *Service) DeleteLink(ctx context.Context, id int64) (LinkView, error) {
	link, err := s.linkWriter.DeleteLink(ctx, id)
	if err != nil {
		return LinkView{}, fmt.Errorf("delete link: %w", err)
	}

	return link, nil
}

// saveLinkFields validates the fields shared by create and update and persists
// them through persist, which differs only in the writer call per operation.
func (s *Service) saveLinkFields(
	ctx context.Context,
	originalURL string,
	shortName string,
	persist func(ctx context.Context, normalizedURL, name string) (LinkView, error),
) (LinkView, error) {
	normalized, err := normalizeURL(originalURL)
	if err != nil {
		return LinkView{}, &FieldError{
			Field: fieldOriginalURL,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidURL, originalURL),
		}
	}

	if shortName == "" {
		return s.withGeneratedShortName(ctx, func(ctx context.Context, name string) (LinkView, error) {
			return persist(ctx, normalized, name)
		})
	}

	if _, err := domainlinks.NewShortCode(shortName); err != nil {
		return LinkView{}, &FieldError{
			Field: fieldShortName,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidShortCode, shortName),
		}
	}

	link, err := persist(ctx, normalized, shortName)
	if err != nil {
		return LinkView{}, fmt.Errorf("persist link: %w", err)
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
	try func(ctx context.Context, shortName string) (LinkView, error),
) (LinkView, error) {
	for attempt := 0; attempt < maxShortCodeAttempts; attempt++ {
		shortName, err := s.generateShortCode()
		if err != nil {
			return LinkView{}, fmt.Errorf("generate short code: %w", err)
		}
		if _, err := domainlinks.NewShortCode(shortName); err != nil {
			continue
		}

		link, err := try(ctx, shortName)
		if err == nil {
			return link, nil
		}

		var fieldErr *FieldError
		if !errors.As(err, &fieldErr) || !errors.Is(fieldErr.Err, ErrShortNameAlreadyUse) {
			return LinkView{}, fmt.Errorf("persist link: %w", err)
		}
	}

	return LinkView{}, fmt.Errorf(
		"%w: exhausted %d attempts",
		ErrShortCodeGenerationFailed,
		maxShortCodeAttempts,
	)
}
