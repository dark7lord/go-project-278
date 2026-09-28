package httpadapter

import (
	"context"

	"github.com/stretchr/testify/mock"

	"code/internal/application"
)

type mockLinkUseCase struct{ mock.Mock }

func (m *mockLinkUseCase) CreateLink(
	ctx context.Context,
	cmd application.CreateLinkCommand,
) (application.LinkView, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(application.LinkView), args.Error(1)
}

func (m *mockLinkUseCase) Redirect(ctx context.Context, cmd application.RedirectCommand) (application.LinkView, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(application.LinkView), args.Error(1)
}

func (m *mockLinkUseCase) GetLinkByID(ctx context.Context, id int64) (application.LinkView, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(application.LinkView), args.Error(1)
}

func (m *mockLinkUseCase) GetLinkByShortName(ctx context.Context, shortName string) (application.LinkView, error) {
	args := m.Called(ctx, shortName)
	return args.Get(0).(application.LinkView), args.Error(1)
}

func (m *mockLinkUseCase) PageLinks(
	ctx context.Context,
	q application.ListLinksQuery,
) (application.RangePage[application.LinkView], error) {
	args := m.Called(ctx, q)
	return args.Get(0).(application.RangePage[application.LinkView]), args.Error(1)
}

func (m *mockLinkUseCase) UpdateLink(
	ctx context.Context,
	id int64,
	cmd application.UpdateLinkCommand,
) (application.LinkView, error) {
	args := m.Called(ctx, id, cmd)
	return args.Get(0).(application.LinkView), args.Error(1)
}

func (m *mockLinkUseCase) DeleteLink(ctx context.Context, id int64) (application.LinkView, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(application.LinkView), args.Error(1)
}

type mockVisitUseCase struct{ mock.Mock }

func (m *mockVisitUseCase) PageLinkVisits(
	ctx context.Context,
	q application.ListLinkVisitsQuery,
) (application.RangePage[application.VisitView], error) {
	args := m.Called(ctx, q)
	return args.Get(0).(application.RangePage[application.VisitView]), args.Error(1)
}
