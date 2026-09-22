package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceRedirectRecordsVisit(t *testing.T) {
	referer := testExampleURL
	link := LinkView{
		ID:          42,
		OriginalURL: testExampleURL,
		ShortName:   testTargetName,
	}
	reader := &fakeLinkReader{gotLink: link}
	recorder := &mockVisitRecorder{}
	svc := NewServiceWithGenerator(
		serviceDeps(reader, nil, nil, recorder),
		&fakeShortCodeGenerator{value: testShortCode},
	)
	recorder.
		On("CreateLinkVisit", mock.Anything, int64(42), "192.0.2.1", "test-agent", &referer, int32(302)).
		Return(VisitView{}, nil).
		Once()

	result, err := svc.Redirect(context.Background(), RedirectCommand{
		ShortName: testTargetName,
		VisitMeta: VisitMeta{
			IP:        "192.0.2.1",
			UserAgent: "test-agent",
			Referer:   &referer,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, link, result)
	recorder.AssertExpectations(t)
}

func TestServiceRedirectRejectsInvalidStoredURL(t *testing.T) {
	recorder := &mockVisitRecorder{}
	svc := NewServiceWithGenerator(
		serviceDeps(
			&fakeLinkReader{gotLink: LinkView{ID: 1, OriginalURL: "ftp://example.com"}},
			nil,
			nil,
			recorder,
		),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	_, err := svc.Redirect(context.Background(), RedirectCommand{ShortName: testTargetName})

	assert.Error(t, err)
	m := mock.Anything
	recorder.AssertNotCalled(t, "CreateLinkVisit", m, m, m, m, m, m)
}

func TestServiceRedirectReturnsVisitError(t *testing.T) {
	visitErr := errors.New("record failed")
	reader := &fakeLinkReader{
		gotLink: LinkView{ID: 1, OriginalURL: testExampleURL},
	}
	recorder := &mockVisitRecorder{}
	svc := NewServiceWithGenerator(
		serviceDeps(reader, nil, nil, recorder),
		&fakeShortCodeGenerator{value: testShortCode},
	)
	recorder.
		On("CreateLinkVisit", mock.Anything, int64(1), "", "", (*string)(nil), int32(302)).
		Return(VisitView{}, visitErr).
		Once()

	_, err := svc.Redirect(context.Background(), RedirectCommand{ShortName: testTargetName})

	assert.ErrorIs(t, err, visitErr)
	recorder.AssertExpectations(t)
}
