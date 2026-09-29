package application

import (
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
	recorder := NewMockVisitRecorder(t)
	svc := NewServiceWithGenerator(
		serviceDeps(reader, nil, nil, recorder),
		&fakeShortCodeGenerator{value: testShortCode},
	)
	recorder.EXPECT().
		CreateLinkVisit(mock.Anything, int64(42), "192.0.2.1", "test-agent", &referer, testRedirectStatus).
		Return(VisitView{}, nil).
		Once()

	result, err := svc.Redirect(t.Context(), RedirectCommand{
		ShortName: testTargetName,
		VisitMeta: VisitMeta{
			IP:        "192.0.2.1",
			UserAgent: "test-agent",
			Referer:   &referer,
		},
		Status: testRedirectStatus,
	})

	require.NoError(t, err)
	assert.Equal(t, link, result)
}

func TestServiceRedirectRejectsInvalidStoredURL(t *testing.T) {
	recorder := NewMockVisitRecorder(t)
	svc := NewServiceWithGenerator(
		serviceDeps(
			&fakeLinkReader{gotLink: LinkView{ID: 1, OriginalURL: "ftp://example.com"}},
			nil,
			nil,
			recorder,
		),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	_, err := svc.Redirect(t.Context(), RedirectCommand{ShortName: testTargetName})

	assert.Error(t, err)
	m := mock.Anything
	recorder.AssertNotCalled(t, "CreateLinkVisit", m, m, m, m, m, m)
}

func TestServiceRedirectReturnsVisitError(t *testing.T) {
	visitErr := errors.New("record failed")
	reader := &fakeLinkReader{
		gotLink: LinkView{ID: 1, OriginalURL: testExampleURL},
	}
	recorder := NewMockVisitRecorder(t)
	svc := NewServiceWithGenerator(
		serviceDeps(reader, nil, nil, recorder),
		&fakeShortCodeGenerator{value: testShortCode},
	)
	recorder.EXPECT().
		CreateLinkVisit(mock.Anything, int64(1), "", "", (*string)(nil), testRedirectStatus).
		Return(VisitView{}, visitErr).
		Once()

	_, err := svc.Redirect(t.Context(), RedirectCommand{
		ShortName: testTargetName,
		Status:    testRedirectStatus,
	})

	assert.ErrorIs(t, err, visitErr)
}
