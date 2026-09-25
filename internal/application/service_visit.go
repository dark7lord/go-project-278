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

// ListLinkVisitsRange retrieves a paginated subset of link visits.
func (s *Service) ListLinkVisitsRange(ctx context.Context, q ListLinkVisitsQuery) (RangePage[VisitView], error) {
	totalVisits, err := s.visitReader.CountLinkVisits(ctx)
	if err != nil {
		return RangePage[VisitView]{}, fmt.Errorf("count link visits: %w", err)
	}

	visits, err := s.visitReader.ListLinkVisitsRange(ctx, q.Start, q.End, q.Sort)
	if err != nil {
		return RangePage[VisitView]{}, fmt.Errorf("list link visits range: %w", err)
	}
	if visits == nil {
		visits = []VisitView{}
	}

	return RangePage[VisitView]{Items: visits, Start: q.Start, Total: totalVisits}, nil
}
