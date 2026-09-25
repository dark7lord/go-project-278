// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"

	"code/internal/application"
	"code/internal/db"
)

// Visit sortable field names matching the ones validated in the HTTP layer.
const (
	fieldLinkID    = application.SortFieldLinkID
	fieldCreatedAt = application.SortFieldCreatedAt
	fieldIP        = application.SortFieldIP
	fieldUserAgent = application.SortFieldUserAgent
	fieldReferer   = application.SortFieldReferer
	fieldStatus    = application.SortFieldStatus
)

// ListLinkVisitsRange retrieves a paginated subset of link visits, sorted when a sort is requested.
func (r *VisitRepository) ListLinkVisitsRange(
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
func (r *VisitRepository) pickVisitsRange(
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
