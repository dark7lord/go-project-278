// Package shortcode provides short-code generators for link adapters.
package shortcode

import (
	"fmt"
	"math/rand/v2"
)

var colors = []string{"red", "orange", "yellow", "green", "cyan", "blue", "purple"}

// Generator produces URL-safe short codes for new links.
type Generator struct{}

// NewGenerator creates a generator backed by randomness.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a short code in the format color-color-link-number.
func (g *Generator) Generate() string {
	prefixColor := colors[rand.IntN(len(colors))]
	infixColor := colors[rand.IntN(len(colors))]
	suffixNum := rand.IntN(1024)

	return fmt.Sprintf("%s-%s-link-%d", prefixColor, infixColor, suffixNum)
}
