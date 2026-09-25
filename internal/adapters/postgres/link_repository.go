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

const shortNameConstraint = "links_short_name_key"

// LinkRepository adapts generated SQL queries to the application persistence ports.
type LinkRepository struct {
	queries *db.Queries
}

// Compile-time checks that the adapter implements the application ports.
var _ application.LinkReader = (*LinkRepository)(nil)
var _ application.LinkWriter = (*LinkRepository)(nil)

// NewLinkRepository creates a PostgreSQL link repository.
func NewLinkRepository(queries *db.Queries) *LinkRepository {
	return &LinkRepository{queries: queries}
}

// toLinkView maps a storage link to its application link view.
func toLinkView(link db.Link) application.LinkView {
	return application.LinkView{
		ID:          link.ID,
		ShortName:   link.ShortName,
		OriginalURL: link.OriginalURL,
	}
}

func mapStorageError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == shortNameConstraint {
		return &application.FieldError{
			Field: string(fieldShortName),
			Err:   application.ErrShortNameAlreadyUse,
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

	return toLinkView(link), mapStorageError(err)
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

	return toLinkView(link), mapStorageError(err)
}

// DeleteLink deletes a link by its ID.
func (r *LinkRepository) DeleteLink(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.DeleteLink(ctx, id)

	return toLinkView(link), mapStorageError(err)
}
