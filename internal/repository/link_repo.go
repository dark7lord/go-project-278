// Package repository implements data access layer for links.
package repository

import (
	"context"

	"code/internal/db"
)

// LinkRepository provides data access for links.
type LinkRepository struct {
	queries *db.Queries
}

// NewLinkRepository creates a new LinkRepository.
func NewLinkRepository(queries *db.Queries) *LinkRepository {
	return &LinkRepository{queries: queries}
}

// GetLinkByID retrieves a link by its ID.
func (r *LinkRepository) GetLinkByID(ctx context.Context, id int64) (db.Link, error) {
	return r.queries.GetLink(ctx, id)
}

// ListLinks retrieves all links.
func (r *LinkRepository) ListLinks(ctx context.Context) ([]db.Link, error) {
	return r.queries.GetLinks(ctx)
}

// ListShortNames retrieves all short names.
func (r *LinkRepository) ListShortNames(ctx context.Context) ([]string, error) {
	return r.queries.GetShortNames(ctx)
}

// ListLinksRange retrieves a paginated subset of links.
func (r *LinkRepository) ListLinksRange(ctx context.Context, limit, offset int64) ([]db.Link, error) {
	return r.queries.GetLinksRange(ctx, db.GetLinksRangeParams{
		Limit:  limit,
		Offset: offset,
	})
}

// CountLinks returns the total number of links.
func (r *LinkRepository) CountLinks(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}

// CreateLink inserts a new link.
func (r *LinkRepository) CreateLink(ctx context.Context, originalURL, shortName, shortURL string) (db.Link, error) {
	return r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})
}

// UpdateLink updates an existing link.
func (r *LinkRepository) UpdateLink(ctx context.Context, id int64, originalURL, shortName, shortURL string) (db.Link, error) {
	return r.queries.UpdateLink(ctx, db.UpdateLinkParams{
		ID:          id,
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})
}

// DeleteLink deletes a link by its ID.
func (r *LinkRepository) DeleteLink(ctx context.Context, id int64) error {
	return r.queries.DeleteLink(ctx, id)
}
