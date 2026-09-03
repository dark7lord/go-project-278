// Package postgres contains PostgreSQL adapters for application ports.
package postgres

import (
	"context"

	"code/internal/application"
	"code/internal/db"
)

// LinkRepository adapts generated SQL queries to the application repository port.
type LinkRepository struct {
	queries *db.Queries
}

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

func (r *LinkRepository) GetLinkByID(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.GetLinkByID(ctx, id)
	return toLinkView(link), err
}

func (r *LinkRepository) GetLinkByShortName(ctx context.Context, shortName string) (application.LinkView, error) {
	link, err := r.queries.GetLinkByShortName(ctx, shortName)
	return toLinkView(link), err
}

func (r *LinkRepository) ListLinks(ctx context.Context) ([]application.LinkView, error) {
	links, err := r.queries.GetLinks(ctx)
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = toLinkView(link)
	}
	return views, err
}

func (r *LinkRepository) ListLinksRange(ctx context.Context, limit, offset int64) ([]application.LinkView, error) {
	links, err := r.queries.GetLinksRange(ctx, db.GetLinksRangeParams{Limit: int32(limit), Offset: int32(offset)})
	views := make([]application.LinkView, len(links))
	for index, link := range links {
		views[index] = toLinkView(link)
	}
	return views, err
}

func (r *LinkRepository) CountLinks(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}

func (r *LinkRepository) CreateLink(ctx context.Context, originalURL, shortName, shortURL string) (application.LinkView, error) {
	link, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})
	return toLinkView(link), err
}

func (r *LinkRepository) UpdateLink(ctx context.Context, id int64, originalURL, shortName, shortURL string) (application.LinkView, error) {
	link, err := r.queries.UpdateLink(ctx, db.UpdateLinkParams{
		ID:          id,
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    shortURL,
	})
	return toLinkView(link), err
}

func (r *LinkRepository) DeleteLink(ctx context.Context, id int64) (application.LinkView, error) {
	link, err := r.queries.DeleteLink(ctx, id)
	return toLinkView(link), err
}

func (r *LinkRepository) CreateLinkVisit(ctx context.Context, linkID int64, ip, userAgent string, referer *string, status int32) (application.VisitView, error) {
	visit, err := r.queries.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
		LinkID:    linkID,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
		Status:    status,
	})
	return toVisitView(visit), err
}

func (r *LinkRepository) ListLinkVisits(ctx context.Context) ([]application.VisitView, error) {
	visits, err := r.queries.GetLinkVisits(ctx)
	views := make([]application.VisitView, len(visits))
	for index, visit := range visits {
		views[index] = toVisitView(visit)
	}
	return views, err
}

func (r *LinkRepository) ListLinkVisitsRange(ctx context.Context, limit, offset int64) ([]application.VisitView, error) {
	visits, err := r.queries.GetLinkVisitsRange(ctx, db.GetLinkVisitsRangeParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	views := make([]application.VisitView, len(visits))
	for index, visit := range visits {
		views[index] = toVisitView(visit)
	}
	return views, err
}

func (r *LinkRepository) CountLinkVisits(ctx context.Context) (int64, error) {
	return r.queries.CountLinkVisits(ctx)
}
