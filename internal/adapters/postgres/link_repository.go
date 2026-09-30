// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"
	"fmt"

	"code/db/generated"
	"code/internal/application"
)

// LinkRepository adapts generated SQL queries to the application persistence ports.
type LinkRepository struct {
	queries *db.Queries
}

// Compile-time check that the adapter implements the application port.
var _ application.LinkStore = (*LinkRepository)(nil)

// NewLinkRepository creates a PostgreSQL link repository.
func NewLinkRepository(queries *db.Queries) *LinkRepository {
	return &LinkRepository{queries: queries}
}

// toLink maps a storage link to its application link.
func toLink(link db.Link) application.Link {
	return application.Link{
		ID:          link.ID,
		ShortName:   link.ShortName,
		OriginalURL: link.OriginalURL,
	}
}

// mapSlice converts every item of a slice with convert.
func mapSlice[T, U any](items []T, convert func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = convert(item)
	}

	return result
}

// GetLinkByID retrieves a link by its ID.
func (r *LinkRepository) GetLinkByID(ctx context.Context, id int64) (application.Link, error) {
	link, err := r.queries.GetLinkByID(ctx, id)
	if err != nil {
		return application.Link{}, mapStorageError(err)
	}

	return toLink(link), nil
}

// GetLinkByShortName retrieves a link by its short name.
func (r *LinkRepository) GetLinkByShortName(ctx context.Context, shortName string) (application.Link, error) {
	link, err := r.queries.GetLinkByShortName(ctx, shortName)
	if err != nil {
		return application.Link{}, mapStorageError(err)
	}

	return toLink(link), nil
}

// PageLinks retrieves a paginated page of links together with the total count.
func (r *LinkRepository) PageLinks(
	ctx context.Context,
	q application.PageQuery,
) (application.RangePage[application.Link], error) {
	total, err := r.queries.CountLinks(ctx)
	if err != nil {
		return application.RangePage[application.Link]{}, fmt.Errorf("count links: %w", err)
	}

	first, last, ok := q.Range.Resolve(total)
	if !ok {
		return application.RangePage[application.Link]{Total: total}, nil
	}

	links, err := r.queries.GetLinksPage(ctx, db.GetLinksPageParams{
		SortField: q.Sort.Field,
		SortAsc:   q.Sort.Asc,
		Offset:    first,
		Limit:     last - first + 1,
	})
	if err != nil {
		return application.RangePage[application.Link]{}, fmt.Errorf(
			"list links range: %w",
			mapStorageError(err),
		)
	}

	return application.RangePage[application.Link]{
		Items: mapSlice(links, toLink),
		First: first,
		Total: total,
	}, nil
}

// CreateLink inserts a new link.
func (r *LinkRepository) CreateLink(ctx context.Context, originalURL, shortName string) (application.Link, error) {
	link, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		return application.Link{}, mapStorageError(err)
	}

	return toLink(link), nil
}

// UpdateLink updates an existing link.
func (r *LinkRepository) UpdateLink(
	ctx context.Context,
	id int64,
	originalURL, shortName string,
) (application.Link, error) {
	link, err := r.queries.UpdateLink(ctx, db.UpdateLinkParams{
		ID:          id,
		OriginalURL: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		return application.Link{}, mapStorageError(err)
	}

	return toLink(link), nil
}

// DeleteLink deletes a link by its ID.
func (r *LinkRepository) DeleteLink(ctx context.Context, id int64) (application.Link, error) {
	link, err := r.queries.DeleteLink(ctx, id)
	if err != nil {
		return application.Link{}, mapStorageError(err)
	}

	return toLink(link), nil
}
