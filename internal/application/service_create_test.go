package application

import (
	"context"
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
		command            CreateLinkCommand
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
			command:       CreateLinkCommand{OriginalURL: "example.com", ShortName: testShortName},
			wantURL:       testExampleURL,
			wantCode:      testShortName,
			wantRepoCalls: 1,
		},
		{
			name:          "trims explicit short name",
			command:       CreateLinkCommand{OriginalURL: testExampleURL, ShortName: " my-link "},
			wantURL:       testExampleURL,
			wantCode:      testShortName,
			wantRepoCalls: 1,
		},
		{
			name:               "generates missing short name",
			command:            CreateLinkCommand{OriginalURL: "https://example.com"},
			generator:          testGeneratedCode,
			wantURL:            testExampleURL,
			wantCode:           testGeneratedCode,
			wantRepoCalls:      1,
			wantGeneratorCalls: 1,
		},
		{
			name:      "retries generated code after collision",
			command:   CreateLinkCommand{OriginalURL: testExampleURL},
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
			writer := &mockLinkWriter{}
			generator := &fakeShortCodeGenerator{value: tt.generator}
			svc := NewServiceWithGenerator(
				serviceDeps(nil, writer, nil, nil),
				generator,
			)
			expected := LinkView{
				OriginalURL: tt.wantURL,
				ShortName:   tt.wantCode,
			}
			call := writer.On("CreateLink", mock.Anything, tt.wantURL, tt.wantCode)
			if len(tt.repoErrors) > 0 {
				call.Return(LinkView{}, tt.repoErrors[0]).Once()
				writer.
					On("CreateLink", mock.Anything, tt.wantURL, tt.wantCode).
					Return(expected, nil).
					Once()
			} else {
				call.Return(expected, nil).Once()
			}

			link, err := svc.CreateLink(context.Background(), tt.command)

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, link.OriginalURL)
			assert.Equal(t, tt.wantCode, link.ShortName)
			assert.Equal(t, tt.wantGeneratorCalls, generator.calls)
			writer.AssertExpectations(t)
			assert.Equal(t, tt.wantRepoCalls, len(writer.Calls))
		})
	}
}

func TestServiceCreateLinkRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		command   CreateLinkCommand
		wantField string
	}{
		{
			name:      "invalid URL",
			command:   CreateLinkCommand{OriginalURL: "ftp://example.com", ShortName: "my-link"},
			wantField: fieldOriginalURL,
		},
		{
			name:      "invalid short name",
			command:   CreateLinkCommand{OriginalURL: testExampleURL, ShortName: "bad name"},
			wantField: fieldShortName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &mockLinkWriter{}
			svc := NewServiceWithGenerator(
				serviceDeps(nil, writer, nil, nil),
				&fakeShortCodeGenerator{value: testShortCode},
			)

			_, err := svc.CreateLink(context.Background(), tt.command)

			var fieldErr *FieldError
			require.ErrorAs(t, err, &fieldErr)
			assert.Equal(t, tt.wantField, fieldErr.Field)
			writer.AssertNotCalled(t, "CreateLink", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestServiceCreateLinkReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	writer := &mockLinkWriter{}
	svc := NewServiceWithGenerator(
		serviceDeps(nil, writer, nil, nil),
		&fakeShortCodeGenerator{value: testShortCode},
	)
	writer.
		On("CreateLink", mock.Anything, testExampleURL, testShortName).
		Return(LinkView{}, repoErr).
		Once()

	_, err := svc.CreateLink(context.Background(), CreateLinkCommand{
		OriginalURL: testExampleURL,
		ShortName:   testShortName,
	})

	assert.ErrorIs(t, err, repoErr)
	writer.AssertExpectations(t)
}

func TestServiceCreateLinkExhaustsGeneratedCodes(t *testing.T) {
	writer := &mockLinkWriter{}
	generator := &fakeShortCodeGenerator{value: testGeneratedCode}
	svc := NewServiceWithGenerator(
		serviceDeps(nil, writer, nil, nil),
		generator,
	)
	writer.
		On("CreateLink", mock.Anything, testExampleURL, testGeneratedCode).
		Return(LinkView{}, &FieldError{
			Field: fieldShortName,
			Err:   ErrShortNameAlreadyUse,
		})

	_, err := svc.CreateLink(context.Background(), CreateLinkCommand{
		OriginalURL: testExampleURL,
	})

	assert.ErrorIs(t, err, ErrShortCodeGenerationFailed)
	assert.Equal(t, maxShortCodeAttempts, generator.calls)
	assert.Equal(t, maxShortCodeAttempts, len(writer.Calls))
}

func TestServiceCreateLinkDoesNotRetryNonCollisionFieldError(t *testing.T) {
	writer := &mockLinkWriter{}
	generator := &fakeShortCodeGenerator{value: testGeneratedCode}
	svc := NewServiceWithGenerator(
		serviceDeps(nil, writer, nil, nil),
		generator,
	)
	fieldErr := &FieldError{
		Field: fieldShortName,
		Err:   errors.New("storage rejected the name"),
	}
	writer.
		On("CreateLink", mock.Anything, testExampleURL, testGeneratedCode).
		Return(LinkView{}, fieldErr).
		Once()

	_, err := svc.CreateLink(context.Background(), CreateLinkCommand{
		OriginalURL: testExampleURL,
	})

	assert.ErrorIs(t, err, fieldErr)
	assert.Equal(t, 1, generator.calls)
	assert.Equal(t, 1, len(writer.Calls))
}
