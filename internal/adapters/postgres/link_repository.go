// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"code/internal/application"
	"code/internal/db"
)

const (
	shortNameConstraint = "links_short_name_key"
	maxPageSize         = 1000
)

// pageRange converts an inclusive [start,end] range into a capped LIMIT/OFFSET pair.
func pageRange(start, end int64) (limit, offset int64) {
	offset = start
	if end-start >= maxPageSize {
		return maxPageSize, offset
	}

	return end - start + 1, offset
}

// LinkRepository adapts generated SQL queries to the application persistence ports.
type LinkRepository struct {
	queries *db.Queries
	baseURL string
}

// Compile-time checks that the adapter implements the application ports.
var _ application.LinkReader = (*LinkRepository)(nil)
var _ application.LinkWriter = (*LinkRepository)(nil)
var _ application.VisitReader = (*LinkRepository)(nil)
var _ application.VisitRecorder = (*LinkRepository)(nil)

// NewLinkRepository creates a PostgreSQL link repository.
//
// baseURL is the public origin (scheme + host, no trailing slash) under which
// links are reachable via "/r/<short_name>". It is part of the HTTP-facing
// link derivation, so a single repository owns assembling LinkView.ShortURL.
func NewLinkRepository(queries *db.Queries, baseURL string) *LinkRepository {
	return &LinkRepository{queries: queries, baseURL: baseURL}
}

// toLinkView maps a storage link to its application link view. ShortURL is
// derived from the repository base URL here, in the single place where a link
// row becomes a public view, so it can never be forgotten by a use case.
func (r *LinkRepository) toLinkView(link db.Link) application.LinkView {
	return application.LinkView{
		ID:          link.ID,
		ShortName:   link.ShortName,
		OriginalURL: link.OriginalURL,
		ShortURL:    r.baseURL + "/r/" + link.ShortName,
	}
}

func toVisitView(visit db.LinkVisit) application.VisitView {
	return application.VisitView{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt.Time,
		IP:        visit.IP,
		UserAgent: visit.UserAgent,
		Status:    visit.Status,
	}
}

func mapStorageError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == shortNameConstraint {
		return &application.FieldError{
			Field: "short_name",
			Err:   application.ErrShortNameAlreadyUse,
		}
	}

	return err
}

// GetLinkByID retrieves a link by its ID.
func (r *LinkRepository) GetLinkByID(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.GetLinkByID(ctx, id)

	return r.toLinkView(link), mapStorageError(err)
}

// GetLinkByShortName retrieves a link by its short name.
func (r *LinkRepository) GetLinkByShortName(ctx context.Context, shortName string) (application.LinkView, error) {
	link, err := r.queries.GetLinkByShortName(ctx, shortName)

	return r.toLinkView(link), mapStorageError(err)
}

// ListLinks retrieves all links.
func (r *LinkRepository) ListLinks(ctx context.Context) ([]application.LinkView, error) {
	links, err := r.queries.GetLinks(ctx)
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = r.toLinkView(link)
	}

	return views, mapStorageError(err)
}

// ListLinksRange retrieves a paginated subset of links.
func (r *LinkRepository) ListLinksRange(ctx context.Context, start, end int64) ([]application.LinkView, error) {
	limit, offset := pageRange(start, end)
	links, err := r.queries.GetLinksRange(ctx, db.GetLinksRangeParams{Limit: limit, Offset: offset})
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = r.toLinkView(link)
	}

	return views, mapStorageError(err)
}

// CountLinks returns the total number of links.
func (r *LinkRepository) CountLinks(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}

// CreateLink inserts a new link.
func (r *LinkRepository) CreateLink(ctx context.Context, originalURL, shortName string) (application.LinkView, error) {
	link, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
	})

	return r.toLinkView(link), mapStorageError(err)
}

// UpdateLink updates an existing link.
func (r *LinkRepository) UpdateLink(
	ctx context.Context,
	id int64,
	originalURL, shortName string,
) (application.LinkView, error) {
	link, err := r.queries.UpdateLink(ctx, db.UpdateLinkParams{
		ID:          id,
		OriginalURL: originalURL,
		ShortName:   shortName,
	})

	return r.toLinkView(link), mapStorageError(err)
}

// DeleteLink deletes a link by its ID.
func (r *LinkRepository) DeleteLink(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.DeleteLink(ctx, id)

	return r.toLinkView(link), mapStorageError(err)
}

// CreateLinkVisit records a visit for the given link.
func (r *LinkRepository) CreateLinkVisit(
	ctx context.Context,
	linkID int64,
	ip, userAgent string,
	referer *string,
	status int32,
) (application.VisitView, error) {
	visit, err := r.queries.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
		LinkID:    linkID,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
		Status:    status,
	})

	return toVisitView(visit), mapStorageError(err)
}

// ListLinkVisits retrieves all link visits.
func (r *LinkRepository) ListLinkVisits(ctx context.Context) ([]application.VisitView, error) {
	visits, err := r.queries.GetLinkVisits(ctx)
	views := make([]application.VisitView, len(visits))
	for index, visit := range visits {
		views[index] = toVisitView(visit)
	}

	return views, mapStorageError(err)
}

// ListLinkVisitsRange retrieves a paginated subset of link visits.
func (r *LinkRepository) ListLinkVisitsRange(ctx context.Context, start, end int64) ([]application.VisitView, error) {
	limit, offset := pageRange(start, end)
	visits, err := r.queries.GetLinkVisitsRange(ctx, db.GetLinkVisitsRangeParams{
		Limit:  limit,
		Offset: offset,
	})
	views := make([]application.VisitView, len(visits))
	for index, visit := range visits {
		views[index] = toVisitView(visit)
	}

	return views, mapStorageError(err)
}

// CountLinkVisits returns the total number of link visits.
func (r *LinkRepository) CountLinkVisits(ctx context.Context) (int64, error) {
	return r.queries.CountLinkVisits(ctx)
}
