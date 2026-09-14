package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceListVisitsReturnsEmptySlice(t *testing.T) {
	svc := NewServiceWithGenerator(serviceDeps(nil, nil, &fakeVisitReader{}, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	visits, err := svc.ListLinkVisits(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, visits)
	assert.Empty(t, visits)
}

func TestServiceListVisitsRange(t *testing.T) {
	expected := []VisitView{{ID: 3, LinkID: 1}}
	reader := &fakeVisitReader{visitCount: 10, visitsRange: expected}
	svc := NewServiceWithGenerator(serviceDeps(nil, nil, reader, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	page, err := svc.ListLinkVisitsRange(context.Background(), ListLinkVisitsQuery{Start: 5, End: 9})

	require.NoError(t, err)
	assert.Equal(t, expected, page.Items)
	assert.Equal(t, int64(10), page.Total)
	assert.Equal(t, int64(5), page.Start)
}

func TestServiceListVisitsRangeNormalizesNilItems(t *testing.T) {
	reader := &fakeVisitReader{visitCount: 10}
	svc := NewServiceWithGenerator(serviceDeps(nil, nil, reader, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	page, err := svc.ListLinkVisitsRange(context.Background(), ListLinkVisitsQuery{Start: 0, End: 4})

	require.NoError(t, err)
	assert.NotNil(t, page.Items)
	assert.Empty(t, page.Items)
}

func TestServiceListVisitsRangePropagatesCountError(t *testing.T) {
	repoErr := errors.New("count visits failed")
	svc := NewServiceWithGenerator(serviceDeps(nil, nil, &fakeVisitReader{visitCountError: repoErr}, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	_, err := svc.ListLinkVisitsRange(context.Background(), ListLinkVisitsQuery{Start: 0, End: 4})

	assert.ErrorIs(t, err, repoErr)
}
