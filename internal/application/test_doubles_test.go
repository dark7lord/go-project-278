package application

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type fakeLinkReader struct {
	gotLink        LinkView
	getLinkError   error
	links          []LinkView
	linksError     error
	linkRange      []LinkView
	linkRangeError error
	linkCount      int64
	linkCountError error
}

func (f *fakeLinkReader) GetLinkByID(_ context.Context, _ int64) (LinkView, error) {
	return f.gotLink, f.getLinkError
}

func (f *fakeLinkReader) GetLinkByShortName(_ context.Context, _ string) (LinkView, error) {
	return f.gotLink, f.getLinkError
}

func (f *fakeLinkReader) ListLinks(_ context.Context) ([]LinkView, error) {
	return f.links, f.linksError
}

func (f *fakeLinkReader) ListLinksRange(_ context.Context, _, _ int64) ([]LinkView, error) {
	return f.linkRange, f.linkRangeError
}

func (f *fakeLinkReader) CountLinks(_ context.Context) (int64, error) {
	return f.linkCount, f.linkCountError
}

type mockLinkWriter struct{ mock.Mock }

func (m *mockLinkWriter) CreateLink(ctx context.Context, originalURL, shortName, shortURL string) (LinkView, error) {
	args := m.Called(ctx, originalURL, shortName, shortURL)
	var link LinkView
	if value := args.Get(0); value != nil {
		link = value.(LinkView)
	}

	return link, args.Error(1)
}

func (m *mockLinkWriter) UpdateLink(ctx context.Context, id int64, originalURL, shortName, shortURL string) (LinkView, error) {
	args := m.Called(ctx, id, originalURL, shortName, shortURL)
	var link LinkView
	if value := args.Get(0); value != nil {
		link = value.(LinkView)
	}

	return link, args.Error(1)
}

func (m *mockLinkWriter) DeleteLink(ctx context.Context, id int64) (LinkView, error) {
	args := m.Called(ctx, id)
	var link LinkView
	if value := args.Get(0); value != nil {
		link = value.(LinkView)
	}

	return link, args.Error(1)
}

type fakeVisitReader struct {
	visits           []VisitView
	visitsRange      []VisitView
	visitsRangeError error
	visitCount       int64
	visitCountError  error
}

func (f *fakeVisitReader) ListLinkVisits(_ context.Context) ([]VisitView, error) {
	return f.visits, nil
}

func (f *fakeVisitReader) ListLinkVisitsRange(_ context.Context, _, _ int64) ([]VisitView, error) {
	return f.visitsRange, f.visitsRangeError
}

func (f *fakeVisitReader) CountLinkVisits(_ context.Context) (int64, error) {
	if f.visitCountError != nil {
		return 0, f.visitCountError
	}
	if f.visitCount != 0 {
		return f.visitCount, nil
	}

	return int64(len(f.visits)), nil
}

type mockVisitRecorder struct{ mock.Mock }

func (m *mockVisitRecorder) CreateLinkVisit(ctx context.Context, linkID int64, ip, userAgent string, referer *string, status int32) (VisitView, error) {
	args := m.Called(ctx, linkID, ip, userAgent, referer, status)
	var visit VisitView
	if value := args.Get(0); value != nil {
		visit = value.(VisitView)
	}

	return visit, args.Error(1)
}

func serviceDeps(linkReader LinkReader, linkWriter LinkWriter, visitReader VisitReader, visitRecorder VisitRecorder) ServiceDeps {
	return ServiceDeps{
		LinkReader:    linkReader,
		LinkWriter:    linkWriter,
		VisitReader:   visitReader,
		VisitRecorder: visitRecorder,
	}
}
