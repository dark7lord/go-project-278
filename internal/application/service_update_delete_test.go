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
	svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	updated, err := svc.UpdateLink(context.Background(), 7, UpdateLinkCommand{
		OriginalURL: "https://updated.com",
		ShortName:   "updated-link",
	})

	require.NoError(t, err)
	assert.Equal(t, expected, updated)
	writer.AssertExpectations(t)
}

func TestServiceUpdateLinkRejectsInvalidURL(t *testing.T) {
	writer := &mockLinkWriter{}
	svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	_, err := svc.UpdateLink(context.Background(), 7, UpdateLinkCommand{
		OriginalURL: "ftp://bad",
		ShortName:   "ok",
	})

	var fieldErr *FieldError
	require.ErrorAs(t, err, &fieldErr)
	assert.Equal(t, fieldOriginalURL, fieldErr.Field)
	writer.AssertNotCalled(t, "UpdateLink", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestServiceUpdateLinkGeneratesShortName(t *testing.T) {
	writer := &mockLinkWriter{}
	generator := &fakeShortCodeGenerator{value: "test-code"}
	expected := LinkView{ID: 7, OriginalURL: "https://ok.com", ShortName: "test-code"}
	writer.On("UpdateLink", mock.Anything, int64(7), "https://ok.com", "test-code", "http://localhost:8080/r/test-code").
		Return(expected, nil).Once()
	svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil), "http://localhost:8080", generator)

	updated, err := svc.UpdateLink(context.Background(), 7, UpdateLinkCommand{
		OriginalURL: "https://ok.com",
		ShortName:   "",
	})

	require.NoError(t, err)
	assert.Equal(t, expected, updated)
	assert.Equal(t, 1, generator.calls)
	writer.AssertExpectations(t)
}

func TestServiceDeleteLink(t *testing.T) {
	writer := &mockLinkWriter{}
	expected := LinkView{ID: 7, ShortName: testTargetName}
	writer.On("DeleteLink", mock.Anything, int64(7)).Return(expected, nil).Once()
	svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	deleted, err := svc.DeleteLink(context.Background(), 7)

	require.NoError(t, err)
	assert.Equal(t, expected, deleted)
	writer.AssertExpectations(t)
}

func TestServiceDeleteLinkWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("delete failed")
	writer := &mockLinkWriter{}
	writer.On("DeleteLink", mock.Anything, int64(1)).Return(LinkView{}, repoErr).Once()
	svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil), "http://localhost:8080", &fakeShortCodeGenerator{value: testShortCode})

	_, err := svc.DeleteLink(context.Background(), 1)

	assert.ErrorIs(t, err, repoErr)
	writer.AssertExpectations(t)
}
