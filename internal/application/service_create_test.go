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
)

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
			svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil), "http://localhost:8080", generator)
			expected := LinkView{OriginalURL: tt.wantURL, ShortName: tt.wantCode, ShortURL: "http://localhost:8080/r/" + tt.wantCode}
			call := writer.On("CreateLink", mock.Anything, tt.wantURL, tt.wantCode, expected.ShortURL)
			if len(tt.repoErrors) > 0 {
				call.Return(LinkView{}, tt.repoErrors[0]).Once()
				writer.On("CreateLink", mock.Anything, tt.wantURL, tt.wantCode, expected.ShortURL).
					Return(expected, nil).Once()
			} else {
				call.Return(expected, nil).Once()
			}

			link, err := svc.CreateLink(context.Background(), tt.command)

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, link.OriginalURL)
			assert.Equal(t, tt.wantCode, link.ShortName)
			assert.Equal(t, "http://localhost:8080/r/"+tt.wantCode, link.ShortURL)
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
			svc := NewService(serviceDeps(nil, writer, nil, nil), "http://localhost:8080")

			_, err := svc.CreateLink(context.Background(), tt.command)

			var fieldErr *FieldError
			require.ErrorAs(t, err, &fieldErr)
			assert.Equal(t, tt.wantField, fieldErr.Field)
			writer.AssertNotCalled(t, "CreateLink", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestServiceCreateLinkReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	writer := &mockLinkWriter{}
	svc := NewService(serviceDeps(nil, writer, nil, nil), "http://localhost:8080")
	writer.On("CreateLink", mock.Anything, testExampleURL, testShortName, "http://localhost:8080/r/"+testShortName).
		Return(LinkView{}, repoErr).Once()

	_, err := svc.CreateLink(context.Background(), CreateLinkCommand{
		OriginalURL: testExampleURL,
		ShortName:   testShortName,
	})

	assert.ErrorIs(t, err, repoErr)
	writer.AssertExpectations(t)
}
