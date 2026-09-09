package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceReadLinks(t *testing.T) {
	expected := []LinkView{{ID: 1, ShortName: "first"}}
	svc := NewService(serviceDeps(&fakeLinkReader{links: expected}, nil, nil, nil), "http://localhost:8080")

	links, err := svc.ListLinks(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, links)
}

func TestServiceReadLinksReturnsEmptySlice(t *testing.T) {
	svc := NewService(serviceDeps(&fakeLinkReader{}, nil, nil, nil), "http://localhost:8080")

	links, err := svc.ListLinks(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, links)
	assert.Empty(t, links)
}

func TestServiceReadLinksRange(t *testing.T) {
	expected := []LinkView{{ID: 2, ShortName: "second"}}
	reader := &fakeLinkReader{linkCount: 10, linkRange: expected}
	svc := NewService(serviceDeps(reader, nil, nil, nil), "http://localhost:8080")

	links, total, err := svc.ListLinksRange(context.Background(), ListLinksQuery{Start: 5, End: 9})

	require.NoError(t, err)
	assert.Equal(t, expected, links)
	assert.Equal(t, int64(10), total)
}

func TestServiceReadLinksRangePropagatesCountError(t *testing.T) {
	repoErr := errors.New("count failed")
	svc := NewService(serviceDeps(&fakeLinkReader{linkCountError: repoErr}, nil, nil, nil), "http://localhost:8080")

	_, _, err := svc.ListLinksRange(context.Background(), ListLinksQuery{Start: 0, End: 4})

	assert.ErrorIs(t, err, repoErr)
}

func TestServiceGetLink(t *testing.T) {
	expected := LinkView{ID: 7, ShortName: testTargetName}
	svc := NewService(serviceDeps(&fakeLinkReader{gotLink: expected}, nil, nil, nil), "http://localhost:8080")

	link, err := svc.GetLinkByID(context.Background(), expected.ID)

	require.NoError(t, err)
	assert.Equal(t, expected, link)
}

func TestServiceGetLinkWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("read failed")
	svc := NewService(serviceDeps(&fakeLinkReader{getLinkError: repoErr}, nil, nil, nil), "http://localhost:8080")

	_, err := svc.GetLinkByShortName(context.Background(), testTargetName)

	assert.ErrorIs(t, err, repoErr)
}
