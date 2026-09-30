package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServicePageLinkVisits(t *testing.T) {
	query := PageQuery{Range: Range{First: 5, Last: 9}}
	expected := RangePage[Visit]{
		Items: []Visit{{ID: 3, LinkID: 1}},
		First: 5,
		Total: 10,
	}
	visits := NewMockVisitStore(t)
	visits.EXPECT().PageLinkVisits(mock.Anything, query).Return(expected, nil).Once()
	svc := NewService(nil, visits, fixedCode)

	page, err := svc.PageLinkVisits(t.Context(), query)

	require.NoError(t, err)
	assert.Equal(t, expected, page)
}

func TestServicePageLinkVisitsPropagatesStoreError(t *testing.T) {
	repoErr := errors.New("page visits failed")
	visits := NewMockVisitStore(t)
	visits.EXPECT().PageLinkVisits(mock.Anything, mock.Anything).Return(RangePage[Visit]{}, repoErr).Once()
	svc := NewService(nil, visits, fixedCode)

	_, err := svc.PageLinkVisits(t.Context(), PageQuery{Range: Range{First: 0, Last: 4}})

	assert.ErrorIs(t, err, repoErr)
}
