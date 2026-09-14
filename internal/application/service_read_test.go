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
	svc := NewServiceWithGenerator(serviceDeps(&fakeLinkReader{links: expected}, nil, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	links, err := svc.ListLinks(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, links)
}

func TestServiceReadLinksReturnsEmptySlice(t *testing.T) {
	svc := NewServiceWithGenerator(serviceDeps(&fakeLinkReader{}, nil, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	links, err := svc.ListLinks(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, links)
	assert.Empty(t, links)
}

func TestServiceReadLinksRange(t *testing.T) {
	expected := []LinkView{{ID: 2, ShortName: "second"}}
	reader := &fakeLinkReader{linkCount: 10, linkRange: expected}
	svc := NewServiceWithGenerator(serviceDeps(reader, nil, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	page, err := svc.ListLinksRange(context.Background(), ListLinksQuery{Start: 5, End: 9})

	require.NoError(t, err)
	assert.Equal(t, expected, page.Items)
	assert.Equal(t, int64(10), page.Total)
	assert.Equal(t, int64(5), page.Start)
}

func TestServiceReadLinksRangeNormalizesNilItems(t *testing.T) {
	reader := &fakeLinkReader{linkCount: 10}
	svc := NewServiceWithGenerator(serviceDeps(reader, nil, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	page, err := svc.ListLinksRange(context.Background(), ListLinksQuery{Start: 0, End: 4})

	require.NoError(t, err)
	assert.NotNil(t, page.Items)
	assert.Empty(t, page.Items)
}

func TestServiceReadLinksRangePropagatesCountError(t *testing.T) {
	repoErr := errors.New("count failed")
	svc := NewServiceWithGenerator(serviceDeps(&fakeLinkReader{linkCountError: repoErr}, nil, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	_, err := svc.ListLinksRange(context.Background(), ListLinksQuery{Start: 0, End: 4})

	assert.ErrorIs(t, err, repoErr)
}

func TestServiceGetLink(t *testing.T) {
	expected := LinkView{ID: 7, ShortName: testTargetName}
	svc := NewServiceWithGenerator(serviceDeps(&fakeLinkReader{gotLink: expected}, nil, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	link, err := svc.GetLinkByID(context.Background(), expected.ID)

	require.NoError(t, err)
	assert.Equal(t, expected, link)
}

func TestServiceGetLinkWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("read failed")
	svc := NewServiceWithGenerator(serviceDeps(&fakeLinkReader{getLinkError: repoErr}, nil, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	_, err := svc.GetLinkByShortName(context.Background(), testTargetName)

	assert.ErrorIs(t, err, repoErr)
}
