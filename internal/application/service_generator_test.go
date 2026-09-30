package application

import (
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
	writer := NewMockLinkWriter(t)
	generator := &fakeShortCodeGenerator{value: testGeneratedCode}

	svc := NewServiceWithGenerator(serviceDeps(nil, writer, nil, nil),
		generator)
	writer.EXPECT().
		CreateLink(mock.Anything, testExampleURL, testGeneratedCode).
		Return(LinkView{
			OriginalURL: testExampleURL,
			ShortName:   testGeneratedCode,
		}, nil).
		Once()
	link, err := svc.CreateLink(t.Context(), LinkInput{
		OriginalURL: testExampleURL,
		ShortName:   "",
	})

	require.NoError(t, err)
	assert.Equal(t, testGeneratedCode, link.ShortName)
	assert.Equal(t, 1, generator.calls)
}
