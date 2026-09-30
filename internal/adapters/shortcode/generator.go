// Package shortcode generates short names for links.
package shortcode

import (
	"fmt"
	"math/rand/v2"
	"strings"
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

// Generate returns a random short code "<word>-<tail>": word is 2-3
// pronounceable syllables, tail is six decimal digits. A short code is not a
// secret, so math/rand/v2 (seeded from the OS) is enough.
func Generate() string {
	var word strings.Builder
	for range 2 + rand.IntN(2) {
		word.WriteString(syllables[rand.IntN(len(syllables))])
	}

	return fmt.Sprintf("%s-%06d", word.String(), rand.IntN(1_000_000))
}
