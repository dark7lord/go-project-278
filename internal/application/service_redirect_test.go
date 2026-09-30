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
	visit := VisitInput{
		IP:        "192.0.2.1",
		UserAgent: "test-agent",
		Referer:   &referer,
		Status:    testRedirectStatus,
	}
	link := Link{
		ID:          42,
		OriginalURL: testExampleURL,
		ShortName:   testTargetName,
	}
	links := NewMockLinkStore(t)
	links.EXPECT().GetLinkByShortName(mock.Anything, testTargetName).Return(link, nil).Once()
	recorder := NewMockVisitStore(t)
	svc := NewService(links, recorder, fixedCode)
	recorder.EXPECT().
		CreateLinkVisit(mock.Anything, int64(42), visit).
		Return(Visit{}, nil).
		Once()

	result, err := svc.Redirect(t.Context(), testTargetName, visit)

	require.NoError(t, err)
	assert.Equal(t, link, result)
}

func TestServiceRedirectReturnsVisitError(t *testing.T) {
	visitErr := errors.New("record failed")
	links := NewMockLinkStore(t)
	links.EXPECT().
		GetLinkByShortName(mock.Anything, testTargetName).
		Return(Link{ID: 1, OriginalURL: testExampleURL}, nil).
		Once()
	recorder := NewMockVisitStore(t)
	svc := NewService(links, recorder, fixedCode)
	recorder.EXPECT().
		CreateLinkVisit(mock.Anything, int64(1), VisitInput{Status: testRedirectStatus}).
		Return(Visit{}, visitErr).
		Once()

	_, err := svc.Redirect(t.Context(), testTargetName, VisitInput{Status: testRedirectStatus})

	assert.ErrorIs(t, err, visitErr)
}
