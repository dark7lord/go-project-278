package application

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testExampleURL    = "https://example.com"
	testShortName     = "my-link"
	testGeneratedCode = "generated-code"
	testTargetName    = "target"
	testOKURL         = "https://ok.com"
)

// testRedirectStatus is deliberately not 302, so the redirect assertions prove
// that the service records the status it was handed instead of a literal 302.
const testRedirectStatus = int32(307)

func TestServiceCreateLink(t *testing.T) {
	tests := []struct {
		name               string
		command            LinkInput
		generator          string
		repoErrors         []error
		wantURL            string
		wantCode           string
		wantErr            error
		wantField          string
		wantRepoCalls      int
		wantGeneratorCalls int
	}{
		{
			name:          "normalizes explicit link",
			command:       LinkInput{OriginalURL: "example.com", ShortName: testShortName},
			wantURL:       testExampleURL,
			wantCode:      testShortName,
			wantRepoCalls: 1,
		},
		{
			name:          "trims explicit short name",
			command:       LinkInput{OriginalURL: testExampleURL, ShortName: " my-link "},
			wantURL:       testExampleURL,
			wantCode:      testShortName,
			wantRepoCalls: 1,
		},
		{
			name:               "generates missing short name",
			command:            LinkInput{OriginalURL: "https://example.com"},
			generator:          testGeneratedCode,
			wantURL:            testExampleURL,
			wantCode:           testGeneratedCode,
			wantRepoCalls:      1,
			wantGeneratorCalls: 1,
		},
		{
			name:      "retries generated code after collision",
			command:   LinkInput{OriginalURL: testExampleURL},
			generator: testGeneratedCode,
			repoErrors: []error{
				&FieldError{Field: fieldShortName, Err: ErrShortNameAlreadyUse},
			},
			wantURL:            testExampleURL,
			wantCode:           testGeneratedCode,
			wantRepoCalls:      2,
			wantGeneratorCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewMockLinkStore(t)
			generator := &fakeGenerator{value: tt.generator}
			svc := NewService(writer, nil, generator.Generate)
			expected := LinkView{
				OriginalURL: tt.wantURL,
				ShortName:   tt.wantCode,
			}
			call := writer.EXPECT().CreateLink(mock.Anything, tt.wantURL, tt.wantCode)
			if len(tt.repoErrors) > 0 {
				call.Return(LinkView{}, tt.repoErrors[0]).Once()
				writer.EXPECT().
					CreateLink(mock.Anything, tt.wantURL, tt.wantCode).
					Return(expected, nil).
					Once()
			} else {
				call.Return(expected, nil).Once()
			}

			link, err := svc.CreateLink(t.Context(), tt.command)

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, link.OriginalURL)
			assert.Equal(t, tt.wantCode, link.ShortName)
			assert.Equal(t, tt.wantGeneratorCalls, generator.calls)
			assert.Equal(t, tt.wantRepoCalls, len(writer.Calls))
		})
	}
}

func TestServiceCreateLinkRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		command   LinkInput
		wantField string
	}{
		{
			name:      "invalid URL",
			command:   LinkInput{OriginalURL: "ftp://example.com", ShortName: "my-link"},
			wantField: fieldOriginalURL,
		},
		{
			name:      "invalid short name",
			command:   LinkInput{OriginalURL: testExampleURL, ShortName: "bad name"},
			wantField: fieldShortName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewMockLinkStore(t)
			svc := NewService(writer, nil, fixedCode)

			_, err := svc.CreateLink(t.Context(), tt.command)

			var fieldErr *FieldError
			require.ErrorAs(t, err, &fieldErr)
			assert.Equal(t, tt.wantField, fieldErr.Field)
			writer.AssertNotCalled(t, "CreateLink", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestServiceCreateLinkReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	writer := NewMockLinkStore(t)
	svc := NewService(writer, nil, fixedCode)
	writer.EXPECT().
		CreateLink(mock.Anything, testExampleURL, testShortName).
		Return(LinkView{}, repoErr).
		Once()

	_, err := svc.CreateLink(t.Context(), LinkInput{
		OriginalURL: testExampleURL,
		ShortName:   testShortName,
	})

	assert.ErrorIs(t, err, repoErr)
}

func TestServiceCreateLinkExhaustsGeneratedCodes(t *testing.T) {
	writer := NewMockLinkStore(t)
	generator := &fakeGenerator{value: testGeneratedCode}
	svc := NewService(writer, nil, generator.Generate)
	writer.EXPECT().
		CreateLink(mock.Anything, testExampleURL, testGeneratedCode).
		Return(LinkView{}, &FieldError{
			Field: fieldShortName,
			Err:   ErrShortNameAlreadyUse,
		})

	_, err := svc.CreateLink(t.Context(), LinkInput{
		OriginalURL: testExampleURL,
	})

	assert.ErrorIs(t, err, ErrShortCodeGenerationFailed)
	assert.Equal(t, maxShortCodeAttempts, generator.calls)
	assert.Equal(t, maxShortCodeAttempts, len(writer.Calls))
}

func TestServiceCreateLinkDoesNotRetryNonCollisionFieldError(t *testing.T) {
	writer := NewMockLinkStore(t)
	generator := &fakeGenerator{value: testGeneratedCode}
	svc := NewService(writer, nil, generator.Generate)
	fieldErr := &FieldError{
		Field: fieldShortName,
		Err:   errors.New("storage rejected the name"),
	}
	writer.EXPECT().
		CreateLink(mock.Anything, testExampleURL, testGeneratedCode).
		Return(LinkView{}, fieldErr).
		Once()

	_, err := svc.CreateLink(t.Context(), LinkInput{
		OriginalURL: testExampleURL,
	})

	assert.ErrorIs(t, err, fieldErr)
	assert.Equal(t, 1, generator.calls)
	assert.Equal(t, 1, len(writer.Calls))
}
