package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicePageLinkVisits(t *testing.T) {
	expected := []VisitView{{ID: 3, LinkID: 1}}
	reader := &fakeVisitReader{
		visitPage: RangePage[VisitView]{
			Items: expected,
			First: 5,
			Total: 10,
		},
	}
	svc := NewServiceWithGenerator(
		serviceDeps(nil, nil, reader, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	page, err := svc.PageLinkVisits(context.Background(), ListLinkVisitsQuery{Range: Range{First: 5, Last: 9}})

	require.NoError(t, err)
	assert.Equal(t, expected, page.Items)
	assert.Equal(t, int64(10), page.Total)
	assert.Equal(t, int64(5), page.First)
}

func TestServicePageLinkVisitsPropagatesReaderError(t *testing.T) {
	repoErr := errors.New("page visits failed")
	svc := NewServiceWithGenerator(
		serviceDeps(nil, nil, &fakeVisitReader{visitPageError: repoErr}, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	_, err := svc.PageLinkVisits(context.Background(), ListLinkVisitsQuery{Range: Range{First: 0, Last: 4}})

	assert.ErrorIs(t, err, repoErr)
}
