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
	shortURLConstraint  = "links_short_url_key"
)

// LinkRepository adapts generated SQL queries to the application persistence ports.
type LinkRepository struct {
	queries *db.Queries
}

// Compile-time checks that the adapter implements the application ports.
var _ application.LinkReader = (*LinkRepository)(nil)
var _ application.LinkWriter = (*LinkRepository)(nil)
var _ application.VisitReader = (*LinkRepository)(nil)
var _ application.VisitRecorder = (*LinkRepository)(nil)

// NewLinkRepository creates a PostgreSQL link repository.
func NewLinkRepository(queries *db.Queries) *LinkRepository {
	return &LinkRepository{queries: queries}
}

func toLinkView(link db.Link) application.LinkView {
	return application.LinkView{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
		ShortURL:    link.ShortURL,
	}
}

func toVisitView(visit db.LinkVisit) application.VisitView {
	return application.VisitView{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt.Time,
		IP:        visit.IP,
		UserAgent: visit.UserAgent,
		Referer:   visit.Referer,
		Status:    visit.Status,
	}
}

func mapStorageError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case shortNameConstraint:
			return &application.FieldError{Field: "short_name", Err: application.ErrShortNameAlreadyUse}
		case shortURLConstraint:
			return &application.FieldError{Field: "short_url", Err: application.ErrShortURLAlreadyUse}
		}
	}

	return err
}

// GetLinkByID retrieves a link by its ID.
func (r *LinkRepository) GetLinkByID(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.GetLinkByID(ctx, id)

	return toLinkView(link), mapStorageError(err)
}

// GetLinkByShortName retrieves a link by its short name.
func (r *LinkRepository) GetLinkByShortName(ctx context.Context, shortName string) (application.LinkView, error) {
	link, err := r.queries.GetLinkByShortName(ctx, shortName)

	return toLinkView(link), mapStorageError(err)
}

// ListLinks retrieves all links.
func (r *LinkRepository) ListLinks(ctx context.Context) ([]application.LinkView, error) {
	links, err := r.queries.GetLinks(ctx)
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = toLinkView(link)
	}

	return views, mapStorageError(err)
}

// ListLinksRange retrieves a paginated subset of links.
func (r *LinkRepository) ListLinksRange(ctx context.Context, limit, offset int64) ([]application.LinkView, error) {
	links, err := r.queries.GetLinksRange(ctx, db.GetLinksRangeParams{Limit: int32(limit), Offset: int32(offset)})
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = toLinkView(link)
	}

	return views, mapStorageError(err)
}

// CountLinks returns the total number of links.
func (r *LinkRepository) CountLinks(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}

// CreateLink inserts a new link.
func (r *LinkRepository) CreateLink(ctx context.Context, originalURL, shortName, shortURL string) (application.LinkView, error) {
	link, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})

	return toLinkView(link), mapStorageError(err)
}

// UpdateLink updates an existing link.
func (r *LinkRepository) UpdateLink(ctx context.Context, id int64, originalURL, shortName, shortURL string) (application.LinkView, error) {
	link, err := r.queries.UpdateLink(ctx, db.UpdateLinkParams{
		ID:          id,
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})

	return toLinkView(link), mapStorageError(err)
}

// DeleteLink deletes a link by its ID.
func (r *LinkRepository) DeleteLink(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.DeleteLink(ctx, id)

	return toLinkView(link), mapStorageError(err)
}

// CreateLinkVisit records a visit for the given link.
func (r *LinkRepository) CreateLinkVisit(ctx context.Context, linkID int64, ip, userAgent string, referer *string, status int32) (application.VisitView, error) {
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
func (r *LinkRepository) ListLinkVisitsRange(ctx context.Context, limit, offset int64) ([]application.VisitView, error) {
	visits, err := r.queries.GetLinkVisitsRange(ctx, db.GetLinkVisitsRangeParams{
		Limit:  int32(limit),
		Offset: int32(offset),
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
