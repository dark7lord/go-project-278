package httpadapter

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"code/internal/application"
)

// rangeSpecRe matches a single RFC 9110 range-spec: first-last, first- or -suffix.
var rangeSpecRe = regexp.MustCompile(`^(\d*)-(\d*)$`)

// parseRangeHeader parses an RFC 9110 Range header: unit=first-last,
// unit=first- or unit=-suffix. It reports false, so the header is ignored,
// for another unit, a malformed value or several ranges.
func parseRangeHeader(header, unit string) (application.Range, bool) {
	gotUnit, spec, found := strings.Cut(header, "=")
	if !found || !strings.EqualFold(gotUnit, unit) {
		return application.Range{}, false
	}

	// Several ranges carry a comma, which never matches a single range-spec.
	matches := rangeSpecRe.FindStringSubmatch(spec)
	if matches == nil || matches[1] == "" && matches[2] == "" {
		return application.Range{}, false
	}

	if matches[1] == "" {
		length, err := strconv.ParseInt(matches[2], 10, 64)
		if err != nil {
			return application.Range{}, false
		}

		return application.Range{Suffix: true, Length: length}, true
	}

	first, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return application.Range{}, false
	}
	if matches[2] == "" {
		return application.Range{First: first, Last: math.MaxInt64}, true
	}

	last, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil || last < first {
		return application.Range{}, false
	}

	return application.Range{First: first, Last: last}, true
}
