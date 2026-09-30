package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceUpdateLink(t *testing.T) {
	writer := NewMockLinkStore(t)
	expected := Link{
		ID:          7,
		OriginalURL: "https://updated.com",
		ShortName:   "updated-link",
	}
	writer.EXPECT().
		UpdateLink(mock.Anything, int64(7), "https://updated.com", "updated-link").
		Return(expected, nil).
		Once()
	svc := NewService(writer, nil, fixedCode)

	updated, err := svc.UpdateLink(t.Context(), 7, LinkInput{
		OriginalURL: "https://updated.com",
		ShortName:   "updated-link",
	})

	require.NoError(t, err)
	assert.Equal(t, expected, updated)
}

func TestServiceUpdateLinkTrimsShortName(t *testing.T) {
	writer := NewMockLinkStore(t)
	expected := Link{
		ID:          7,
		OriginalURL: "https://updated.com",
		ShortName:   "updated-link",
	}
	writer.EXPECT().
		UpdateLink(mock.Anything, int64(7), "https://updated.com", "updated-link").
		Return(expected, nil).
		Once()
	svc := NewService(writer, nil, fixedCode)

	updated, err := svc.UpdateLink(t.Context(), 7, LinkInput{
		OriginalURL: "https://updated.com",
		ShortName:   " updated-link ",
	})

	require.NoError(t, err)
	assert.Equal(t, expected, updated)
}

func TestServiceUpdateLinkRejectsInvalidURL(t *testing.T) {
	writer := NewMockLinkStore(t)
	svc := NewService(writer, nil, fixedCode)

	_, err := svc.UpdateLink(t.Context(), 7, LinkInput{
		OriginalURL: "ftp://bad",
		ShortName:   "ok",
	})

	var fieldErr *FieldError
	require.ErrorAs(t, err, &fieldErr)
	assert.Equal(t, fieldOriginalURL, fieldErr.Field)
	writer.AssertNotCalled(t, "UpdateLink", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestServiceUpdateLinkGeneratesShortName(t *testing.T) {
	writer := NewMockLinkStore(t)
	generator := &fakeGenerator{value: "test-code"}
	expected := Link{
		ID:          7,
		OriginalURL: testOKURL,
		ShortName:   "test-code",
	}
	writer.EXPECT().
		UpdateLink(mock.Anything, int64(7), testOKURL, "test-code").
		Return(expected, nil).
		Once()
	svc := NewService(writer, nil, generator.Generate)

	updated, err := svc.UpdateLink(t.Context(), 7, LinkInput{
		OriginalURL: testOKURL,
		ShortName:   "",
	})

	require.NoError(t, err)
	assert.Equal(t, expected, updated)
	assert.Equal(t, 1, generator.calls)
}

func TestServiceUpdateLinkExhaustsGeneratedCodes(t *testing.T) {
	writer := NewMockLinkStore(t)
	generator := &fakeGenerator{value: testGeneratedCode}
	svc := NewService(writer, nil, generator.Generate)
	writer.EXPECT().
		UpdateLink(mock.Anything, int64(7), testOKURL, testGeneratedCode).
		Return(Link{}, &FieldError{Field: fieldShortName, Err: ErrShortNameAlreadyUse})

	_, err := svc.UpdateLink(t.Context(), 7, LinkInput{
		OriginalURL: testOKURL,
		ShortName:   "",
	})

	assert.ErrorIs(t, err, ErrShortCodeGenerationFailed)
	assert.Equal(t, maxShortCodeAttempts, generator.calls)
	assert.Equal(t, maxShortCodeAttempts, len(writer.Calls))
}

func TestServiceDeleteLink(t *testing.T) {
	writer := NewMockLinkStore(t)
	expected := Link{
		ID:        7,
		ShortName: testTargetName,
	}
	writer.EXPECT().
		DeleteLink(mock.Anything, int64(7)).
		Return(expected, nil).
		Once()
	svc := NewService(writer, nil, fixedCode)

	deleted, err := svc.DeleteLink(t.Context(), 7)

	require.NoError(t, err)
	assert.Equal(t, expected, deleted)
}

func TestServiceDeleteLinkWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("delete failed")
	writer := NewMockLinkStore(t)
	writer.EXPECT().
		DeleteLink(mock.Anything, int64(1)).
		Return(Link{}, repoErr).
		Once()
	svc := NewService(writer, nil, fixedCode)

	_, err := svc.DeleteLink(t.Context(), 1)

	assert.ErrorIs(t, err, repoErr)
}
