// Package application contains application-level ports and use-case contracts.
package application

import (
	"context"
)

// LinkStore persists links.
type LinkStore interface {
	GetLinkByID(ctx context.Context, id int64) (Link, error)
	GetLinkByShortName(ctx context.Context, shortName string) (Link, error)
	PageLinks(ctx context.Context, q PageQuery) (RangePage[Link], error)
	CreateLink(ctx context.Context, originalURL, shortName string) (Link, error)
	UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (Link, error)
	DeleteLink(ctx context.Context, id int64) (Link, error)
}

// VisitStore persists link visits.
type VisitStore interface {
	CreateLinkVisit(ctx context.Context, linkID int64, visit VisitInput) (Visit, error)
	PageLinkVisits(ctx context.Context, q PageQuery) (RangePage[Visit], error)
}

// UseCase lists the operations a transport, such as the HTTP adapter, calls.
type UseCase interface {
	CreateLink(ctx context.Context, in LinkInput) (Link, error)
	Redirect(ctx context.Context, shortName string, visit VisitInput) (Link, error)
	GetLinkByID(ctx context.Context, id int64) (Link, error)
	GetLinkByShortName(ctx context.Context, shortName string) (Link, error)
	PageLinks(ctx context.Context, q PageQuery) (RangePage[Link], error)
	UpdateLink(ctx context.Context, id int64, in LinkInput) (Link, error)
	DeleteLink(ctx context.Context, id int64) (Link, error)
	PageLinkVisits(ctx context.Context, q PageQuery) (RangePage[Visit], error)
}
