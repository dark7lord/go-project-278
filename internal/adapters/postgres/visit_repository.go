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

// Compile-time checks that the adapter implements the application ports.
var _ application.VisitReader = (*VisitRepository)(nil)
var _ application.VisitRecorder = (*VisitRepository)(nil)

// NewVisitRepository creates a PostgreSQL visit repository.
func NewVisitRepository(queries *db.Queries) *VisitRepository {
	return &VisitRepository{queries: queries}
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

func toVisitViews(visits []db.LinkVisit) []application.VisitView {
	views := make([]application.VisitView, len(visits))
	for index, visit := range visits {
		views[index] = toVisitView(visit)
	}

	return views
}

// CreateLinkVisit records a visit for the given link.
func (r *VisitRepository) CreateLinkVisit(
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
	if err != nil {
		return application.VisitView{}, mapStorageError(err)
	}

	return toVisitView(visit), nil
}

// PageLinkVisits retrieves a paginated page of link visits together with the total count.
func (r *VisitRepository) PageLinkVisits(
	ctx context.Context,
	q application.ListLinkVisitsQuery,
) (application.RangePage[application.VisitView], error) {
	total, err := r.queries.CountLinkVisits(ctx)
	if err != nil {
		return application.RangePage[application.VisitView]{}, fmt.Errorf("count link visits: %w", err)
	}

	first, last, ok := q.Range.Resolve(total)
	if !ok {
		return application.RangePage[application.VisitView]{Total: total}, nil
	}

	visits, err := r.pickVisitsRange(ctx, q.Sort, last-first+1, first)
	if err != nil {
		return application.RangePage[application.VisitView]{}, fmt.Errorf(
			"list link visits range: %w",
			mapStorageError(err),
		)
	}

	return application.RangePage[application.VisitView]{
		Items: toVisitViews(visits),
		First: first,
		Total: total,
	}, nil
}
