// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"

	"code/internal/application"
	"code/internal/db"
)

// Sortable field names matching the ones validated in the HTTP layer.
const (
	fieldID          = application.SortFieldID
	fieldShortName   = application.SortFieldShortName
	fieldOriginalURL = application.SortFieldOriginalURL
	fieldLinkID      = application.SortFieldLinkID
	fieldCreatedAt   = application.SortFieldCreatedAt
	fieldIP          = application.SortFieldIP
	fieldUserAgent   = application.SortFieldUserAgent
	fieldReferer     = application.SortFieldReferer
	fieldStatus      = application.SortFieldStatus
)

// pageRange converts an inclusive [start,end] range into a LIMIT/OFFSET pair.
func pageRange(start, end int64) (limit, offset int64) {
	return end - start + 1, start
}

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

// ListLinkVisitsRange retrieves a paginated subset of link visits, sorted when a sort is requested.
func (r *LinkRepository) ListLinkVisitsRange(
	ctx context.Context,
	start, end int64,
	sort *application.Sort,
) ([]application.VisitView, error) {
	limit, offset := pageRange(start, end)

	visits, err := r.pickVisitsRange(ctx, sort, limit, offset)
	views := make([]application.VisitView, len(visits))
	for index, visit := range visits {
		views[index] = toVisitView(visit)
	}

	return views, mapStorageError(err)
}

// pickVisitsRange selects the query matching a link visits sort request.
func (r *LinkRepository) pickVisitsRange(
	ctx context.Context,
	sort *application.Sort,
	limit, offset int64,
) ([]db.LinkVisit, error) {
	switch {
	case sort == nil || (sort.Field == fieldID && sort.Asc):
		return r.queries.GetLinkVisitsRange(ctx, db.GetLinkVisitsRangeParams{Limit: limit, Offset: offset})
	case sort.Field == fieldID:
		return r.queries.GetLinkVisitsRangeIdDesc(
			ctx,
			db.GetLinkVisitsRangeIdDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldCreatedAt && sort.Asc:
		return r.queries.GetLinkVisitsRangeCreatedAtAsc(
			ctx,
			db.GetLinkVisitsRangeCreatedAtAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldCreatedAt:
		return r.queries.GetLinkVisitsRangeCreatedAtDesc(
			ctx,
			db.GetLinkVisitsRangeCreatedAtDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldLinkID && sort.Asc:
		return r.queries.GetLinkVisitsRangeLinkIdAsc(
			ctx,
			db.GetLinkVisitsRangeLinkIdAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldLinkID:
		return r.queries.GetLinkVisitsRangeLinkIdDesc(
			ctx,
			db.GetLinkVisitsRangeLinkIdDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldIP && sort.Asc:
		return r.queries.GetLinkVisitsRangeIpAsc(
			ctx,
			db.GetLinkVisitsRangeIpAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldIP:
		return r.queries.GetLinkVisitsRangeIpDesc(
			ctx,
			db.GetLinkVisitsRangeIpDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldUserAgent && sort.Asc:
		return r.queries.GetLinkVisitsRangeUserAgentAsc(
			ctx,
			db.GetLinkVisitsRangeUserAgentAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldUserAgent:
		return r.queries.GetLinkVisitsRangeUserAgentDesc(
			ctx,
			db.GetLinkVisitsRangeUserAgentDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldStatus && sort.Asc:
		return r.queries.GetLinkVisitsRangeStatusAsc(
			ctx,
			db.GetLinkVisitsRangeStatusAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldStatus:
		return r.queries.GetLinkVisitsRangeStatusDesc(
			ctx,
			db.GetLinkVisitsRangeStatusDescParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldReferer && sort.Asc:
		return r.queries.GetLinkVisitsRangeRefererAsc(
			ctx,
			db.GetLinkVisitsRangeRefererAscParams{Limit: limit, Offset: offset},
		)
	case sort.Field == fieldReferer:
		return r.queries.GetLinkVisitsRangeRefererDesc(
			ctx,
			db.GetLinkVisitsRangeRefererDescParams{Limit: limit, Offset: offset},
		)
	default:
		return nil, application.ErrSortField
	}
}
