package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceListVisitsReturnsEmptySlice(t *testing.T) {
	svc := NewServiceWithGenerator(
		serviceDeps(nil, nil, &fakeVisitReader{}, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	visits, err := svc.ListLinkVisits(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, visits)
	assert.Empty(t, visits)
}

func TestServicePageLinkVisits(t *testing.T) {
	expected := []VisitView{{ID: 3, LinkID: 1}}
	reader := &fakeVisitReader{
		visitPage: RangePage[VisitView]{
			Items: expected,
			Start: 5,
			Total: 10,
		},
	}
	svc := NewServiceWithGenerator(
		serviceDeps(nil, nil, reader, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	page, err := svc.PageLinkVisits(context.Background(), ListLinkVisitsQuery{Start: 5, End: 9})

	require.NoError(t, err)
	assert.Equal(t, expected, page.Items)
	assert.Equal(t, int64(10), page.Total)
	assert.Equal(t, int64(5), page.Start)
}

func TestServicePageLinkVisitsNormalizesNilItems(t *testing.T) {
	reader := &fakeVisitReader{}
	svc := NewServiceWithGenerator(
		serviceDeps(nil, nil, reader, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	page, err := svc.PageLinkVisits(context.Background(), ListLinkVisitsQuery{Start: 0, End: 4})

	require.NoError(t, err)
	assert.NotNil(t, page.Items)
	assert.Empty(t, page.Items)
}

func TestServicePageLinkVisitsPropagatesReaderError(t *testing.T) {
	repoErr := errors.New("page visits failed")
	svc := NewServiceWithGenerator(
		serviceDeps(nil, nil, &fakeVisitReader{visitPageError: repoErr}, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	_, err := svc.PageLinkVisits(context.Background(), ListLinkVisitsQuery{Start: 0, End: 4})

	assert.ErrorIs(t, err, repoErr)
}
