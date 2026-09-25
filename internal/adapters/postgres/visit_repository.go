// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"

	"code/internal/application"
	"code/internal/db"
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
		Status:    visit.Status,
	}
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

	return toVisitView(visit), mapStorageError(err)
}

// ListLinkVisits retrieves all link visits.
func (r *VisitRepository) ListLinkVisits(ctx context.Context) ([]application.VisitView, error) {
	visits, err := r.queries.GetLinkVisits(ctx)
	views := make([]application.VisitView, len(visits))
	for index, visit := range visits {
		views[index] = toVisitView(visit)
	}

	return views, mapStorageError(err)
}

// CountLinkVisits returns the total number of link visits.
func (r *VisitRepository) CountLinkVisits(ctx context.Context) (int64, error) {
	return r.queries.CountLinkVisits(ctx)
}
