package application

import (
	"context"
	"fmt"
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

// ListLinkVisits retrieves all link visits.
func (s *Service) ListLinkVisits(ctx context.Context) ([]VisitView, error) {
	visits, err := s.visitReader.ListLinkVisits(ctx)
	if err != nil {
		return nil, fmt.Errorf("list link visits: %w", err)
	}
	if visits == nil {
		return []VisitView{}, nil
	}

	return visits, nil
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
	if page.Items == nil {
		page.Items = []VisitView{}
	}

	return page, nil
}
