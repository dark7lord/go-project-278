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
	visit := Visit{
		IP:        "192.0.2.1",
		UserAgent: "test-agent",
		Referer:   &referer,
		Status:    testRedirectStatus,
	}
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
		CreateLinkVisit(mock.Anything, int64(42), visit).
		Return(VisitView{}, nil).
		Once()

	result, err := svc.Redirect(t.Context(), testTargetName, visit)

	require.NoError(t, err)
	assert.Equal(t, link, result)
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
		CreateLinkVisit(mock.Anything, int64(1), Visit{Status: testRedirectStatus}).
		Return(VisitView{}, visitErr).
		Once()

	_, err := svc.Redirect(t.Context(), testTargetName, Visit{Status: testRedirectStatus})

	assert.ErrorIs(t, err, visitErr)
}
