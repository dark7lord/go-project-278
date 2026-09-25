// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"

	"code/internal/application"
	"code/internal/db"
)

// Link sortable field names matching the ones validated in the HTTP layer.
const (
	fieldShortName   = application.SortFieldShortName
	fieldOriginalURL = application.SortFieldOriginalURL
)

// ListLinksRange retrieves a paginated subset of links, sorted when a sort is requested.
func (r *LinkRepository) ListLinksRange(
	ctx context.Context,
	start, end int64,
	sort *application.Sort,
) ([]application.LinkView, error) {
	limit, offset := pageRange(start, end)

	links, err := r.pickLinksRange(ctx, sort, limit, offset)
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = toLinkView(link)
	}

	return views, mapStorageError(err)
}

// pickLinksRange selects the query matching a links sort request.
func (r *LinkRepository) pickLinksRange(
	ctx context.Context,
	sort *application.Sort,
	limit, offset int64,
) ([]db.Link, error) {
	switch {
	case sort == nil || (sort.Field == fieldID && sort.Asc):
		return r.queries.GetLinksRange(ctx, db.GetLinksRangeParams{Limit: limit, Offset: offset})
	case sort.Field == fieldID:
		return r.queries.GetLinksRangeIdDesc(
			ctx,
			db.GetLinksRangeIdDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldShortName && sort.Asc:
		return r.queries.GetLinksRangeShortNameAsc(
			ctx,
			db.GetLinksRangeShortNameAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldShortName:
		return r.queries.GetLinksRangeShortNameDesc(
			ctx,
			db.GetLinksRangeShortNameDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldOriginalURL && sort.Asc:
		return r.queries.GetLinksRangeOriginalURLAsc(
			ctx,
			db.GetLinksRangeOriginalURLAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldOriginalURL:
		return r.queries.GetLinksRangeOriginalURLDesc(
			ctx,
			db.GetLinksRangeOriginalURLDescParams{Limit: limit, Offset: offset},
		)
	default:
		return nil, application.ErrSortField
	}
}
