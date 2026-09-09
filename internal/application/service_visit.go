package application

import "context"

// CreateLinkVisit records a visit for the given link.
func (s *Service) CreateLinkVisit(ctx context.Context, linkID int64, ip, userAgent string, referer *string, status int32) (VisitView, error) {
	return s.visitRecorder.CreateLinkVisit(ctx, linkID, ip, userAgent, referer, status)
}

// ListLinkVisits retrieves all link visits.
func (s *Service) ListLinkVisits(ctx context.Context) ([]VisitView, error) {
	visits, err := s.visitReader.ListLinkVisits(ctx)
	if err != nil {
		return nil, err
	}
	if visits == nil {
		return []VisitView{}, nil
	}

	return visits, nil
}

// ListLinkVisitsRange retrieves a paginated subset of link visits.
func (s *Service) ListLinkVisitsRange(ctx context.Context, q ListLinkVisitsQuery) ([]VisitView, int64, error) {
	totalVisits, err := s.visitReader.CountLinkVisits(ctx)
	if err != nil {
		return nil, 0, err
	}

	start, end := q.Start, q.End
	visits, err := s.visitReader.ListLinkVisitsRange(ctx, (end-start)+1, start)

	return visits, totalVisits, err
}
