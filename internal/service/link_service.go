// Package service implements business logic for links.
package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"slices"

	"code/internal/db"
	"code/internal/repository"
)

// LinkService implements link business logic.
type LinkService struct {
	repo    *repository.LinkRepository
	baseURL string
}

// NewLinkService creates a new LinkService.
func NewLinkService(repo *repository.LinkRepository, baseURL string) *LinkService {
	return &LinkService{repo, baseURL}
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

func (s *LinkService) validateShortName(ctx context.Context, shortName string) error {
	if shortName == "" {
		return ErrEmptyShortName
	}

	allShortNames, err := s.repo.ListShortNames(ctx)
	if err != nil {
		return err
	}

	if slices.Contains(allShortNames, shortName) {
		return errors.New("short name already exists")
	}

	return nil
}

func (s *LinkService) generateShortLink(ctx context.Context, shortName *string) (string, error) {
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
func (s *LinkService) CreateLink(ctx context.Context, originalURL, shortName string) (db.Link, error) {
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
func (s *LinkService) GetLink(ctx context.Context, id int64) (db.Link, error) {
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
func (s *LinkService) ListLinks(ctx context.Context) ([]db.Link, error) {
	return s.repo.ListLinks(ctx)
}

// UpdateLink updates an existing link.
func (s *LinkService) UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (db.Link, error) {
	// TODO: validation

	_, err := s.repo.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Link{}, ErrNotFound
		}

		return db.Link{}, fmt.Errorf("update link: %w", err)
	}

	shortURL, err := s.generateShortLink(ctx, &shortName)
	if err != nil {
		return db.Link{}, err
	}

	updated, err := s.repo.UpdateLink(ctx, id, originalURL, shortName, shortURL)
	if err != nil {
		return db.Link{}, err
	}

	return updated, nil
}

// DeleteLink deletes a link by its ID.
func (s *LinkService) DeleteLink(ctx context.Context, id int64) error {
	_, err := s.repo.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		return fmt.Errorf("delete link: %w", err)
	}

	return s.repo.DeleteLink(ctx, id)
}
