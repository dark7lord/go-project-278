package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testShortCode = "test-code"

func TestServiceCreateLinkUsesInjectedGenerator(t *testing.T) {
	writer := NewMockLinkStore(t)
	generator := &fakeGenerator{value: testGeneratedCode}

	svc := NewService(writer, nil, generator.Generate)
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
