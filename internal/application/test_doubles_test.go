package application

import (
	"context"
)

type fakeLinkReader struct {
	gotLink       LinkView
	getLinkError  error
	linkPage      RangePage[LinkView]
	linkPageError error
}

func (f *fakeLinkReader) GetLinkByID(_ context.Context, _ int64) (LinkView, error) {
	return f.gotLink, f.getLinkError
}

func (f *fakeLinkReader) GetLinkByShortName(_ context.Context, _ string) (LinkView, error) {
	return f.gotLink, f.getLinkError
}

func (f *fakeLinkReader) PageLinks(_ context.Context, _ ListLinksQuery) (RangePage[LinkView], error) {
	return f.linkPage, f.linkPageError
}

type fakeVisitReader struct {
	visitPage      RangePage[VisitView]
	visitPageError error
}

func (f *fakeVisitReader) PageLinkVisits(
	_ context.Context,
	_ ListLinkVisitsQuery,
) (RangePage[VisitView], error) {
	return f.visitPage, f.visitPageError
}

func serviceDeps(
	linkReader LinkReader,
	linkWriter LinkWriter,
	visitReader VisitReader,
	visitRecorder VisitRecorder,
) ServiceDeps {
	return ServiceDeps{
		LinkReader:    linkReader,
		LinkWriter:    linkWriter,
		VisitReader:   visitReader,
		VisitRecorder: visitRecorder,
	}
}
