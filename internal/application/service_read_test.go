package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServicePageLinks(t *testing.T) {
	query := PageQuery{Range: Range{First: 5, Last: 9}}
	expected := RangePage[Link]{
		Items: []Link{{ID: 2, ShortName: "second"}},
		First: 5,
		Total: 10,
	}
	links := NewMockLinkStore(t)
	links.EXPECT().PageLinks(mock.Anything, query).Return(expected, nil).Once()
	svc := NewService(links, nil, fixedCode)

	page, err := svc.PageLinks(t.Context(), query)

	require.NoError(t, err)
	assert.Equal(t, expected, page)
}

func TestServicePageLinksPropagatesStoreError(t *testing.T) {
	repoErr := errors.New("page failed")
	links := NewMockLinkStore(t)
	links.EXPECT().PageLinks(mock.Anything, mock.Anything).Return(RangePage[Link]{}, repoErr).Once()
	svc := NewService(links, nil, fixedCode)

	_, err := svc.PageLinks(t.Context(), PageQuery{Range: Range{First: 0, Last: 4}})

	assert.ErrorIs(t, err, repoErr)
}

func TestServiceGetLink(t *testing.T) {
	expected := Link{ID: 7, ShortName: testTargetName}
	links := NewMockLinkStore(t)
	links.EXPECT().GetLinkByID(mock.Anything, expected.ID).Return(expected, nil).Once()
	svc := NewService(links, nil, fixedCode)

	link, err := svc.GetLinkByID(t.Context(), expected.ID)

	require.NoError(t, err)
	assert.Equal(t, expected, link)
}

func TestServiceGetLinkWrapsStoreError(t *testing.T) {
	repoErr := errors.New("read failed")
	links := NewMockLinkStore(t)
	links.EXPECT().GetLinkByShortName(mock.Anything, testTargetName).Return(Link{}, repoErr).Once()
	svc := NewService(links, nil, fixedCode)

	_, err := svc.GetLinkByShortName(t.Context(), testTargetName)

	assert.ErrorIs(t, err, repoErr)
}
