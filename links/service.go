package links

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"slices"

	"code/db"
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
func (s *Service) validateShortName(ctx context.Context, shortName string) error {
	if shortName == "" {
		return ErrEmptyShortName
	}

	allShortNames, err := s.repo.ListShortNames(ctx)
	if err != nil {
		return err
	}

	if slices.Contains(allShortNames, shortName) {
		return ErrLinkExists
	}

	return nil
}

// generateShortLink returns the full short URL for the given short name,
// generating a random unique name if the provided one is empty.
func (s *Service) generateShortLink(ctx context.Context, shortName *string) (string, error) {
	err := s.validateShortName(ctx, *shortName)

	if errors.Is(err, ErrEmptyShortName) {
		isValid := false
		for !isValid {
			*shortName = genRandomName()
			if s.validateShortName(ctx, *shortName) == nil {
				isValid = true
			}
		}
	} else if err != nil {
		return "", err
	}

	return s.baseURL + "/" + *shortName, nil
}

// CreateLink creates a new link with the given URL and optional short name.
func (s *Service) CreateLink(ctx context.Context, originalURL, shortName string) (db.Link, error) {
	if shortName != "" {
		if err := s.validateShortName(ctx, shortName); err != nil {
			return db.Link{}, err
		}
	}

	shortURL, err := s.generateShortLink(ctx, &shortName)

	if err != nil {
		return db.Link{}, err
	}

	return s.repo.CreateLink(ctx, originalURL, shortName, shortURL)
}

// GetLink retrieves a link by its ID.
func (s *Service) GetLink(ctx context.Context, id int64) (db.Link, error) {
	link, err := s.repo.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	// TODO: validation

	shortURL, err := s.generateShortLink(ctx, &shortName)
	if err != nil {
		return db.Link{}, err
	}

	updated, err := s.repo.UpdateLink(ctx, id, originalURL, shortName, shortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, fmt.Errorf("delete link: %w", err)
	}

	return link, nil
}
