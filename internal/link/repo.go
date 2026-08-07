// Package link implements HTTP handlers, business logic, and data access for links.
package link

import (
	"context"

	"code/internal/db"
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

// ListLinksRange retrieves a paginated subset of links.
func (r *Repository) ListLinksRange(ctx context.Context, limit, offset int64) ([]db.Link, error) {
	return r.queries.GetLinksRange(ctx, db.GetLinksRangeParams{
		Limit:  int32(limit),
		Offset: int32(offset),
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

// CreateLinkVisit records a visit for the given link.
func (r *Repository) CreateLinkVisit(ctx context.Context, linkID int64, ip, userAgent string, referer *string, status int32) (db.LinkVisit, error) {
	return r.queries.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
		LinkID:    linkID,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
		Status:    status,
	})
}

// ListLinkVisits retrieves all link visits.
func (r *Repository) ListLinkVisits(ctx context.Context) ([]db.LinkVisit, error) {
	return r.queries.GetLinkVisits(ctx)
}

// ListLinkVisitsRange retrieves a paginated subset of link visits.
func (r *Repository) ListLinkVisitsRange(ctx context.Context, limit, offset int64) ([]db.LinkVisit, error) {
	return r.queries.GetLinkVisitsRange(ctx, db.GetLinkVisitsRangeParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
}

// CountLinkVisits returns the total number of link visits.
func (r *Repository) CountLinkVisits(ctx context.Context) (int64, error) {
	return r.queries.CountLinkVisits(ctx)
}
