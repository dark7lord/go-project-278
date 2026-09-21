package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeShortCodeGenerator struct {
	value string
	calls int
}

const testShortCode = "test-code"

func (g *fakeShortCodeGenerator) Generate() string {
	g.calls++
	return g.value
}

func TestServiceCreateLinkUsesInjectedGenerator(t *testing.T) {
	writer := &mockLinkWriter{}
	generator := &fakeShortCodeGenerator{value: testGeneratedCode}

	svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil), "http://localhost:8080", generator)
	writer.
		On("CreateLink", mock.Anything, testExampleURL, testGeneratedCode).
		Return(LinkView{
			OriginalURL: testExampleURL,
			ShortName:   testGeneratedCode,
			ShortURL:    "http://localhost:8080/r/" + testGeneratedCode,
		}, nil).
		Once()
	link, err := svc.CreateLink(context.Background(), CreateLinkCommand{
		OriginalURL: testExampleURL,
		ShortName:   "",
	})

	require.NoError(t, err)
	assert.Equal(t, testGeneratedCode, link.ShortName)
	assert.Equal(t, "http://localhost:8080/r/generated-code", link.ShortURL)
	writer.AssertExpectations(t)
	assert.Equal(t, 1, generator.calls)
}
