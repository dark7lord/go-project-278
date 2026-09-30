package shortcode

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

var generatedCodeRe = regexp.MustCompile(`^[a-z]{4,6}-\d{6}$`)

func TestGenerate(t *testing.T) {
	generator := NewGenerator()
	seen := make(map[string]struct{}, 1000)

	for i := 0; i < 1000; i++ {
		code := generator.Generate()
		assert.Regexp(t, generatedCodeRe, code)
		seen[code] = struct{}{}
	}

	assert.Len(t, seen, 1000)
}
