package application

import (
	"context"
	"errors"
	"fmt"

	domainlinks "code/internal/domain/links"
)

// CreateLink creates a new link with the given URL and optional short name.
func (s *Service) CreateLink(ctx context.Context, cmd CreateLinkCommand) (LinkView, error) {
	normalized, err := normalizeURL(cmd.OriginalURL)
	if err != nil {
		return LinkView{}, &FieldError{
			Field: fieldOriginalURL,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidURL, cmd.OriginalURL),
		}
	}

	shortName := cmd.ShortName
	if shortName == "" {
		for {
			shortName = s.generateShortCode()
			if _, err := domainlinks.NewShortCode(shortName); err != nil {
				continue
			}
			link, err := s.linkWriter.CreateLink(ctx, normalized, shortName, s.generateShortLink(shortName))
			if err == nil {
				return link, nil
			}
			var fieldErr *FieldError
			if !errors.As(err, &fieldErr) {
				return LinkView{}, err
			}
		}
	}

	if _, err := domainlinks.NewShortCode(shortName); err != nil {
		return LinkView{}, &FieldError{
			Field: fieldShortName,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidShortCode, shortName),
		}
	}

	link, err := s.linkWriter.CreateLink(ctx, normalized, shortName, s.generateShortLink(shortName))
	if err != nil {
		return LinkView{}, err
	}

	return link, nil
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

	if _, err := s.CreateLinkVisit(
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
		return nil, err
	}
	if links == nil {
		return []LinkView{}, nil
	}

	return links, nil
}

// ListLinksRange retrieves a paginated subset of links.
func (s *Service) ListLinksRange(ctx context.Context, q ListLinksQuery) ([]LinkView, int64, error) {
	totalLinks, err := s.linkReader.CountLinks(ctx)
	if err != nil {
		return nil, 0, err
	}

	start, end := q.Start, q.End
	links, err := s.linkReader.ListLinksRange(ctx, (end-start)+1, start)

	return links, totalLinks, err
}

// UpdateLink updates an existing link.
func (s *Service) UpdateLink(ctx context.Context, id int64, cmd UpdateLinkCommand) (LinkView, error) {
	normalized, err := normalizeURL(cmd.OriginalURL)
	if err != nil {
		return LinkView{}, &FieldError{
			Field: fieldOriginalURL,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidURL, cmd.OriginalURL),
		}
	}

	shortName := cmd.ShortName
	if shortName == "" {
		for {
			shortName = s.generateShortCode()
			if _, err := domainlinks.NewShortCode(shortName); err != nil {
				continue
			}
			updated, err := s.linkWriter.UpdateLink(ctx, id, normalized, shortName, s.generateShortLink(shortName))
			if err == nil {
				return updated, nil
			}
			var fieldErr *FieldError
			if !errors.As(err, &fieldErr) {
				return LinkView{}, err
			}
		}
	}

	if _, err := domainlinks.NewShortCode(shortName); err != nil {
		return LinkView{}, &FieldError{
			Field: fieldShortName,
			Err:   fmt.Errorf("%w: %s", domainlinks.ErrInvalidShortCode, shortName),
		}
	}

	updated, err := s.linkWriter.UpdateLink(ctx, id, normalized, shortName, s.generateShortLink(shortName))
	if err != nil {
		return LinkView{}, err
	}

	return updated, nil
}

// DeleteLink deletes a link by its ID.
func (s *Service) DeleteLink(ctx context.Context, id int64) (LinkView, error) {
	link, err := s.linkWriter.DeleteLink(ctx, id)
	if err != nil {
		return LinkView{}, fmt.Errorf("delete link: %w", err)
	}

	return link, nil
}
