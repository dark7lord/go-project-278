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

func toLinkViews(links []db.Link) []application.LinkView {
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = toLinkView(link)
	}

	return views
}

// GetLinkByID retrieves a link by its ID.
func (r *LinkRepository) GetLinkByID(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.GetLinkByID(ctx, id)
	if err != nil {
		return application.LinkView{}, mapStorageError(err)
	}

	return toLinkView(link), nil
}

// GetLinkByShortName retrieves a link by its short name.
func (r *LinkRepository) GetLinkByShortName(ctx context.Context, shortName string) (application.LinkView, error) {
	link, err := r.queries.GetLinkByShortName(ctx, shortName)
	if err != nil {
		return application.LinkView{}, mapStorageError(err)
	}

	return toLinkView(link), nil
}

// PageLinks retrieves a paginated page of links together with the total count.
func (r *LinkRepository) PageLinks(
	ctx context.Context,
	q application.ListLinksQuery,
) (application.RangePage[application.LinkView], error) {
	total, err := r.queries.CountLinks(ctx)
	if err != nil {
		return application.RangePage[application.LinkView]{}, fmt.Errorf("count links: %w", err)
	}

	first, last, ok := q.Range.Resolve(total)
	if !ok {
		return application.RangePage[application.LinkView]{Total: total}, nil
	}

	field, asc := sortParams(q.Sort)
	links, err := r.queries.GetLinksPage(ctx, db.GetLinksPageParams{
		SortField: field,
		SortAsc:   asc,
		Offset:    first,
		Limit:     last - first + 1,
	})
	if err != nil {
		return application.RangePage[application.LinkView]{}, fmt.Errorf(
			"list links range: %w",
			mapStorageError(err),
		)
	}

	return application.RangePage[application.LinkView]{
		Items: toLinkViews(links),
		First: first,
		Total: total,
	}, nil
}

// CreateLink inserts a new link.
func (r *LinkRepository) CreateLink(ctx context.Context, originalURL, shortName string) (application.LinkView, error) {
	link, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		return application.LinkView{}, mapStorageError(err)
	}

	return toLinkView(link), nil
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
	if err != nil {
		return application.LinkView{}, mapStorageError(err)
	}

	return toLinkView(link), nil
}

// DeleteLink deletes a link by its ID.
func (r *LinkRepository) DeleteLink(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.DeleteLink(ctx, id)
	if err != nil {
		return application.LinkView{}, mapStorageError(err)
	}

	return toLinkView(link), nil
}
