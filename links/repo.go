// Package links implements HTTP handlers, business logic, and data access for links.
package links

import (
	"context"

	"code/db"
)

// Repository provides data access for links.
type Repository struct {
	queries *db.Queries
}

// NewRepository creates a new Repository.
func NewRepository(queries *db.Queries) *Repository {
	return &Repository{queries: queries}
}

// GetLinkByID retrieves a link by its ID.
func (r *Repository) GetLinkByID(ctx context.Context, id int64) (db.Link, error) {
	return r.queries.GetLinkByID(ctx, id)
}

// GetLinkByShortName retrieves a link by its short name.
func (r *Repository) GetLinkByShortName(ctx context.Context, shortName string) (db.Link, error) {
	return r.queries.GetLinkByShortName(ctx, shortName)
}

// ListLinks retrieves all links.
func (r *Repository) ListLinks(ctx context.Context) ([]db.Link, error) {
	return r.queries.GetLinks(ctx)
}

// ListShortNames retrieves all short names.
func (r *Repository) ListShortNames(ctx context.Context) ([]string, error) {
	return r.queries.GetShortNames(ctx)
}

// ListLinksRange retrieves a paginated subset of links.
func (r *Repository) ListLinksRange(ctx context.Context, limit, offset int64) ([]db.Link, error) {
	return r.queries.GetLinksRange(ctx, db.GetLinksRangeParams{
		Limit:  limit,
		Offset: offset,
	})
}

// CountLinks returns the total number of links.
func (r *Repository) CountLinks(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}

// CreateLink inserts a new link.
func (r *Repository) CreateLink(ctx context.Context, originalURL, shortName, shortURL string) (db.Link, error) {
	return r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})
}

// UpdateLink updates an existing link.
func (r *Repository) UpdateLink(ctx context.Context, id int64, originalURL, shortName, shortURL string) (db.Link, error) {
	return r.queries.UpdateLink(ctx, db.UpdateLinkParams{
		ID:          id,
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})
}

// DeleteLink deletes a link by its ID.
func (r *Repository) DeleteLink(ctx context.Context, id int64) (db.Link, error) {
	return r.queries.DeleteLink(ctx, id)
}
