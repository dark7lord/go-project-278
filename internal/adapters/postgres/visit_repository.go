package postgres

import (
	"context"
	"fmt"

	"code/db/generated"
	"code/internal/application"
)

// VisitRepository adapts generated visit queries to the application persistence ports.
type VisitRepository struct {
	queries *db.Queries
}

// Compile-time check that the adapter implements the application port.
var _ application.VisitStore = (*VisitRepository)(nil)

// NewVisitRepository creates a PostgreSQL visit repository.
func NewVisitRepository(queries *db.Queries) *VisitRepository {
	return &VisitRepository{queries: queries}
}

func toVisit(visit db.LinkVisit) application.Visit {
	return application.Visit{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt.Time,
		IP:        visit.IP,
		UserAgent: visit.UserAgent,
		Referer:   visit.Referer,
		Status:    visit.Status,
	}
}

// CreateLinkVisit records a visit for the given link.
func (r *VisitRepository) CreateLinkVisit(
	ctx context.Context,
	linkID int64,
	visit application.VisitInput,
) (application.Visit, error) {
	created, err := r.queries.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
		LinkID:    linkID,
		IP:        visit.IP,
		UserAgent: visit.UserAgent,
		Referer:   visit.Referer,
		Status:    visit.Status,
	})
	if err != nil {
		return application.Visit{}, mapStorageError(err)
	}

	return toVisit(created), nil
}

// PageLinkVisits retrieves a paginated page of link visits together with the total count.
func (r *VisitRepository) PageLinkVisits(
	ctx context.Context,
	q application.PageQuery,
) (application.RangePage[application.Visit], error) {
	total, err := r.queries.CountLinkVisits(ctx)
	if err != nil {
		return application.RangePage[application.Visit]{}, fmt.Errorf("count link visits: %w", err)
	}

	first, last, ok := q.Range.Resolve(total)
	if !ok {
		return application.RangePage[application.Visit]{Total: total}, nil
	}

	visits, err := r.queries.GetLinkVisitsPage(ctx, db.GetLinkVisitsPageParams{
		SortField: q.Sort.Field,
		SortAsc:   q.Sort.Asc,
		Offset:    first,
		Limit:     last - first + 1,
	})
	if err != nil {
		return application.RangePage[application.Visit]{}, fmt.Errorf(
			"list link visits range: %w",
			mapStorageError(err),
		)
	}

	return application.RangePage[application.Visit]{
		Items: mapSlice(visits, toVisit),
		First: first,
		Total: total,
	}, nil
}
