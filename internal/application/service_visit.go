package application

import (
	"context"
)

// createLinkVisit records a visit for the given link.
func (s *Service) createLinkVisit(
	ctx context.Context,
	linkID int64,
	ip, userAgent string,
	referer *string,
	status int32,
) (VisitView, error) {
	return s.visitRecorder.CreateLinkVisit(ctx, linkID, ip, userAgent, referer, status)
}

// PageLinkVisits retrieves a paginated page of link visits.
func (s *Service) PageLinkVisits(
	ctx context.Context,
	q ListLinkVisitsQuery,
) (RangePage[VisitView], error) {
	page, err := s.visitReader.PageLinkVisits(ctx, q)
	if err != nil {
		return RangePage[VisitView]{}, err
	}

	return page, nil
}
