// Package application contains application-level ports and use-case contracts.
package application

import (
	"context"
)

// LinkReader defines persistence operations for reading links.
type LinkReader interface {
	GetLinkByID(ctx context.Context, id int64) (LinkView, error)
	GetLinkByShortName(ctx context.Context, shortName string) (LinkView, error)
	PageLinks(ctx context.Context, q ListLinksQuery) (RangePage[LinkView], error)
}

// LinkWriter defines persistence operations for changing links.
type LinkWriter interface {
	CreateLink(ctx context.Context, originalURL, shortName string) (LinkView, error)
	UpdateLink(ctx context.Context, id int64, originalURL, shortName string) (LinkView, error)
	DeleteLink(ctx context.Context, id int64) (LinkView, error)
}

// VisitReader defines persistence operations for reading visits.
type VisitReader interface {
	PageLinkVisits(ctx context.Context, q ListLinkVisitsQuery) (RangePage[VisitView], error)
}

// VisitRecorder defines persistence operations for recording visits.
type VisitRecorder interface {
	CreateLinkVisit(
		ctx context.Context,
		linkID int64,
		ip, userAgent string,
		referer *string,
		status int32,
	) (VisitView, error)
}

// ServiceDeps contains the output ports used by the application service.
type ServiceDeps struct {
	LinkReader    LinkReader
	LinkWriter    LinkWriter
	VisitReader   VisitReader
	VisitRecorder VisitRecorder
}

// LinkUseCase defines the link operations required by the HTTP transport.
type LinkUseCase interface {
	CreateLink(ctx context.Context, cmd CreateLinkCommand) (LinkView, error)
	Redirect(ctx context.Context, cmd RedirectCommand) (LinkView, error)
	GetLinkByID(ctx context.Context, id int64) (LinkView, error)
	GetLinkByShortName(ctx context.Context, shortName string) (LinkView, error)
	PageLinks(ctx context.Context, q ListLinksQuery) (RangePage[LinkView], error)
	UpdateLink(ctx context.Context, id int64, cmd UpdateLinkCommand) (LinkView, error)
	DeleteLink(ctx context.Context, id int64) (LinkView, error)
}

// VisitUseCase defines the visit operations required by the HTTP transport.
type VisitUseCase interface {
	PageLinkVisits(ctx context.Context, q ListLinkVisitsQuery) (RangePage[VisitView], error)
}

// UseCase combines the application use cases for a single HTTP adapter.
type UseCase interface {
	LinkUseCase
	VisitUseCase
}

// ShortCodeGenerator creates link identifiers for new links.
type ShortCodeGenerator interface {
	Generate() (string, error)
}
