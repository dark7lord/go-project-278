// Package application contains application-level ports and use-case contracts.
package application

import (
	"context"
)

// LinkRepository defines the persistence operations required by link use cases.
type LinkRepository interface {
	GetLinkByID(ctx context.Context, id int64) (LinkView, error)
	GetLinkByShortName(ctx context.Context, shortName string) (LinkView, error)
	ListLinks(ctx context.Context) ([]LinkView, error)
	ListLinksRange(ctx context.Context, limit, offset int64) ([]LinkView, error)
	CountLinks(ctx context.Context) (int64, error)
	CreateLink(ctx context.Context, originalURL, shortName, shortURL string) (LinkView, error)
	UpdateLink(ctx context.Context, id int64, originalURL, shortName, shortURL string) (LinkView, error)
	DeleteLink(ctx context.Context, id int64) (LinkView, error)
	CreateLinkVisit(ctx context.Context, linkID int64, ip, userAgent string, referer *string, status int32) (VisitView, error)
	ListLinkVisits(ctx context.Context) ([]VisitView, error)
	ListLinkVisitsRange(ctx context.Context, limit, offset int64) ([]VisitView, error)
	CountLinkVisits(ctx context.Context) (int64, error)
}
