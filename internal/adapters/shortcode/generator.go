// Package shortcode provides short-code generators for the link adapters.
package shortcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

const (
	// tailDigits is the length of the decimal tail appended to generated words.
	tailDigits = 6

	// tailPow10 is 10^tailDigits, the upper bound for the decimal tail.
	tailPow10 = 1_000_000
)

// syllables are lowercase pronounceable two-letter chunks used to build
// Docker-style short words such as "vimba-482103".
var syllables = []string{
	"ba", "be", "bi", "bo", "bu",
	"ca", "ce", "ci", "co", "cu",
	"da", "de", "di", "do", "du",
	"fa", "fe", "fi", "fo", "fu",
	"ga", "ge", "gi", "go", "gu",
	"ha", "he", "hi", "ho", "hu",
	"ka", "ke", "ki", "ko", "ku",
	"la", "le", "li", "lo", "lu",
	"ma", "me", "mi", "mo", "mu",
	"na", "ne", "ni", "no", "nu",
	"pa", "pe", "pi", "po", "pu",
	"ra", "re", "ri", "ro", "ru",
	"sa", "se", "si", "so", "su",
	"ta", "te", "ti", "to", "tu",
	"va", "ve", "vi", "vo",
	"wa", "we", "wi", "wo",
	"xa", "xe", "xi", "xo",
	"za", "ze", "zi", "zo",
}

// Generator produces Docker-style short code strings (word-tail).
type Generator struct{}

// NewGenerator creates a short code generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate returns a random short code of the form "<word>-<tail>", where
// word is 2-3 pronounceable syllables and tail is tailDigits decimal digits.
func (g *Generator) Generate() (string, error) {
	word, err := g.randomWord()
	if err != nil {
		return "", err
	}

	tail, err := g.randomTail()
	if err != nil {
		return "", err
	}

	return word + "-" + tail, nil
}

// randomWord builds a 2- or 3-syllable lowercase word.
func (g *Generator) randomWord() (string, error) {
	syllableCount, err := cryptoInt(2)
	if err != nil {
		return "", fmt.Errorf("choose syllable count: %w", err)
	}

	var word strings.Builder
	for i := 0; i < syllableCount+2; i++ {
		index, err := cryptoInt(len(syllables))
		if err != nil {
			return "", fmt.Errorf("choose syllable: %w", err)
		}
		word.WriteString(syllables[index])
	}

	return word.String(), nil
}

// randomTail returns a zero-padded decimal tail of tailDigits digits.
func (g *Generator) randomTail() (string, error) {
	value, err := cryptoInt(tailPow10)
	if err != nil {
		return "", fmt.Errorf("choose tail: %w", err)
	}

	return fmt.Sprintf("%0*d", tailDigits, value), nil
}

// cryptoInt returns a cryptographically uniform integer in [0, n) using
// crypto/rand.Int's built-in rejection sampling (no modulo bias); callers
// pass positive bounds only.
func cryptoInt(n int) (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}

	return int(value.Int64()), nil
}
