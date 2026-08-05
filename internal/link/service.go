package link

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"

	"github.com/jackc/pgx/v5"

	"code/internal/db"
)

// Service implements link business logic.
type Service struct {
	repo    *Repository
	baseURL string
}

// NewService creates a new Service.
func NewService(repo *Repository, baseURL string) *Service {
	return &Service{repo: repo, baseURL: baseURL}
}

var colors = []string{"red", "orange", "yellow", "green", "cyan", "blue", "purple"}

var (
	// ErrNotFound indicates the requested link was not found.
	ErrNotFound = errors.New("link not found")
	// ErrLinkExists indicates a link with the given short name already exists.
	ErrLinkExists = errors.New("link already exists")
	// ErrEmptyShortName indicates the short name is empty.
	ErrEmptyShortName = errors.New("short name is empty")
)

func genRandomName() string {
	l := len(colors)
	prefixColor := colors[rand.Intn(l)]
	infixColor := colors[rand.Intn(l)]
	suffixNum := rand.Intn(1024)

	return fmt.Sprintf("%s-%s-link-%d", prefixColor, infixColor, suffixNum)
}

// validateShortName checks that the short name is non-empty and unique.
// excludeID, if provided, is excluded from the uniqueness check (for updates).
func (s *Service) validateShortName(ctx context.Context, shortName string, excludeID *int64) error {
	if shortName == "" {
		return ErrEmptyShortName
	}

	var allShortNames []string
	var err error
	if excludeID != nil {
		allShortNames, err = s.repo.ListShortNamesExcluding(ctx, *excludeID)
	} else {
		allShortNames, err = s.repo.ListShortNames(ctx)
	}
	if err != nil {
		return err
	}

	if slices.Contains(allShortNames, shortName) {
		return ErrLinkExists
	}

	return nil
}

// generateShortLink returns the full short URL for the given short name.
func (s *Service) generateShortLink(shortName string) string {
	return s.baseURL + "/r/" + shortName
}

// CreateLink creates a new link with the given URL and optional short name.
func (s *Service) CreateLink(ctx context.Context, originalURL, shortName string) (db.Link, error) {
	normalized, err := normalizeURL(originalURL)
	if err != nil {
		return db.Link{}, fmt.Errorf("%w: %s", ErrInvalidURL, originalURL)
	}
	originalURL = normalized

	if shortName == "" {
		for {
			shortName = genRandomName()
			if err := s.validateShortName(ctx, shortName, nil); err == nil {
				break
			}
		}
	} else {
		if err := s.validateShortName(ctx, shortName, nil); err != nil {
			return db.Link{}, err
		}
	}

	shortURL := s.generateShortLink(shortName)

	return s.repo.CreateLink(ctx, originalURL, shortName, shortURL)
}

// GetLinkByID retrieves a link by its ID.
func (s *Service) GetLinkByID(ctx context.Context, id int64) (db.Link, error) {
	link, err := s.repo.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// GetLinkByShortName retrieves a link by its short name.
func (s *Service) GetLinkByShortName(ctx context.Context, shortName string) (db.Link, error) {
	link, err := s.repo.GetLinkByShortName(ctx, shortName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, fmt.Errorf("get link: %w", err)
	}

	return link, nil
}

// ListLinks retrieves all links.
func (s *Service) ListLinks(ctx context.Context) ([]db.Link, error) {
	links, err := s.repo.ListLinks(ctx)
	if err != nil {
		return nil, err
	}
	if links == nil {
		return []db.Link{}, nil
	}

	return links, nil
}

// ListLinksRange retrieves a paginated subset of links.
func (s *Service) ListLinksRange(ctx context.Context, start, end int64) ([]db.Link, int64, error) {
	totalLinks, err := s.repo.CountLinks(ctx)
	if err != nil {
		return nil, 0, err
	}

	limit := (end - start) + 1
	offset := start

	links, err := s.repo.ListLinksRange(ctx, limit, offset)

	return links, totalLinks, err
}

// UpdateLink updates an existing link.
func (s *Service) UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (db.Link, error) {
	normalized, err := normalizeURL(originalURL)
	if err != nil {
		return db.Link{}, fmt.Errorf("%w: %s", ErrInvalidURL, originalURL)
	}
	originalURL = normalized

	if err := s.validateShortName(ctx, shortName, &id); err != nil {
		return db.Link{}, err
	}

	shortURL := s.generateShortLink(shortName)

	updated, err := s.repo.UpdateLink(ctx, id, originalURL, shortName, shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, fmt.Errorf("update link: %w", err)
	}

	return updated, nil
}

// DeleteLink deletes a link by its ID.
func (s *Service) DeleteLink(ctx context.Context, id int64) (db.Link, error) {
	link, err := s.repo.DeleteLink(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, fmt.Errorf("delete link: %w", err)
	}

	return link, nil
}

// CreateLinkVisit records a visit for the given link.
func (s *Service) CreateLinkVisit(ctx context.Context, linkID int64, ip, userAgent string, referer *string, status int32) (db.LinkVisit, error) {
	return s.repo.CreateLinkVisit(ctx, linkID, ip, userAgent, referer, status)
}

// ListLinkVisits retrieves all link visits.
func (s *Service) ListLinkVisits(ctx context.Context) ([]db.LinkVisit, error) {
	visits, err := s.repo.ListLinkVisits(ctx)
	if err != nil {
		return nil, err
	}
	if visits == nil {
		return []db.LinkVisit{}, nil
	}

	return visits, nil
}

// ListLinkVisitsRange retrieves a paginated subset of link visits.
func (s *Service) ListLinkVisitsRange(ctx context.Context, start, end int64) ([]db.LinkVisit, int64, error) {
	totalVisits, err := s.repo.CountLinkVisits(ctx)
	if err != nil {
		return nil, 0, err
	}

	limit := (end - start) + 1
	offset := start

	visits, err := s.repo.ListLinkVisitsRange(ctx, limit, offset)

	return visits, totalVisits, err
}
