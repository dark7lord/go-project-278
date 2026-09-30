// Package application contains application-level ports and use-case contracts.
package application

import (
	"context"
)

// LinkStore persists links.
type LinkStore interface {
	GetLinkByID(ctx context.Context, id int64) (LinkView, error)
	GetLinkByShortName(ctx context.Context, shortName string) (LinkView, error)
	PageLinks(ctx context.Context, q PageQuery) (RangePage[LinkView], error)
	CreateLink(ctx context.Context, originalURL, shortName string) (LinkView, error)
	UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (LinkView, error)
	DeleteLink(ctx context.Context, id int64) (LinkView, error)
}

// VisitStore persists link visits.
type VisitStore interface {
	CreateLinkVisit(ctx context.Context, linkID int64, visit Visit) (VisitView, error)
	PageLinkVisits(ctx context.Context, q PageQuery) (RangePage[VisitView], error)
}

// UseCase lists the operations a transport, such as the HTTP adapter, calls.
type UseCase interface {
	CreateLink(ctx context.Context, in LinkInput) (LinkView, error)
	Redirect(ctx context.Context, shortName string, visit Visit) (LinkView, error)
	GetLinkByID(ctx context.Context, id int64) (LinkView, error)
	GetLinkByShortName(ctx context.Context, shortName string) (LinkView, error)
	PageLinks(ctx context.Context, q PageQuery) (RangePage[LinkView], error)
	UpdateLink(ctx context.Context, id int64, in LinkInput) (LinkView, error)
	DeleteLink(ctx context.Context, id int64) (LinkView, error)
	PageLinkVisits(ctx context.Context, q PageQuery) (RangePage[VisitView], error)
}
