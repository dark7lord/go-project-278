package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServicePageLinks(t *testing.T) {
	expected := []LinkView{{
		ID:        2,
		ShortName: "second",
	}}
	reader := &fakeLinkReader{
		linkPage: RangePage[LinkView]{
			Items: expected,
			First: 5,
			Total: 10,
		},
	}
	svc := NewServiceWithGenerator(
		serviceDeps(reader, nil, nil, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	page, err := svc.PageLinks(context.Background(), ListLinksQuery{Range: Range{First: 5, Last: 9}})

	require.NoError(t, err)
	assert.Equal(t, expected, page.Items)
	assert.Equal(t, int64(10), page.Total)
	assert.Equal(t, int64(5), page.First)
}

func TestServicePageLinksPropagatesReaderError(t *testing.T) {
	repoErr := errors.New("page failed")
	svc := NewServiceWithGenerator(
		serviceDeps(&fakeLinkReader{linkPageError: repoErr}, nil, nil, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	_, err := svc.PageLinks(context.Background(), ListLinksQuery{Range: Range{First: 0, Last: 4}})

	assert.ErrorIs(t, err, repoErr)
}

func TestServiceGetLink(t *testing.T) {
	expected := LinkView{
		ID:        7,
		ShortName: testTargetName,
	}
	svc := NewServiceWithGenerator(
		serviceDeps(&fakeLinkReader{gotLink: expected}, nil, nil, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	link, err := svc.GetLinkByID(context.Background(), expected.ID)

	require.NoError(t, err)
	assert.Equal(t, expected, link)
}

func TestServiceGetLinkWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("read failed")
	svc := NewServiceWithGenerator(
		serviceDeps(&fakeLinkReader{getLinkError: repoErr}, nil, nil, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)

	_, err := svc.GetLinkByShortName(context.Background(), testTargetName)

	assert.ErrorIs(t, err, repoErr)
}
