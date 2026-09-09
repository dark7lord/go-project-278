package application

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceUpdateLink(t *testing.T) {
	writer := &mockLinkWriter{}
	expected := LinkView{ID: 7, OriginalURL: "https://updated.com", ShortName: "updated-link"}
	writer.On("UpdateLink", mock.Anything, int64(7), "https://updated.com", "updated-link", "http://localhost:8080/r/updated-link").
		Return(expected, nil).Once()
	svc := NewService(serviceDeps(nil, writer, nil, nil), "http://localhost:8080")

	updated, err := svc.UpdateLink(context.Background(), 7, UpdateLinkCommand{
		OriginalURL: "https://updated.com",
		ShortName:   "updated-link",
	})

	require.NoError(t, err)
	assert.Equal(t, expected, updated)
	writer.AssertExpectations(t)
}

func TestServiceDeleteLink(t *testing.T) {
	writer := &mockLinkWriter{}
	expected := LinkView{ID: 7, ShortName: testTargetName}
	writer.On("DeleteLink", mock.Anything, int64(7)).Return(expected, nil).Once()
	svc := NewService(serviceDeps(nil, writer, nil, nil), "http://localhost:8080")

	deleted, err := svc.DeleteLink(context.Background(), 7)

	require.NoError(t, err)
	assert.Equal(t, expected, deleted)
	writer.AssertExpectations(t)
}

func TestServiceDeleteLinkWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("delete failed")
	writer := &mockLinkWriter{}
	writer.On("DeleteLink", mock.Anything, int64(1)).Return(LinkView{}, repoErr).Once()
	svc := NewService(serviceDeps(nil, writer, nil, nil), "http://localhost:8080")

	_, err := svc.DeleteLink(context.Background(), 1)

	assert.ErrorIs(t, err, repoErr)
	writer.AssertExpectations(t)
}
