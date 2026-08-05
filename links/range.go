package links

import (
	"errors"
	"regexp"
	"strconv"
)

var rangeRe = regexp.MustCompile(`\[(\d+),(\d+)\]`) // captures: [full, start, end]

var (
	// ErrRangeFormat indicates the range value does not match [start,end].
	ErrRangeFormat = errors.New("invalid range, expected [start,end]")
	// ErrRangeStart indicates the start value is not a number.
	ErrRangeStart = errors.New("invalid start value")
	// ErrRangeEnd indicates the end value is not a number.
	ErrRangeEnd = errors.New("invalid end value")
	// ErrRangeNotSatisfiable indicates the range is negative or inverted.
	ErrRangeNotSatisfiable = errors.New("range not satisfiable")
)

// parseRangeParam parses a "range" query parameter value into start and end.
func parseRangeParam(rangeParam string) (start, end int, err error) {
	matches := rangeRe.FindStringSubmatch(rangeParam)
	if len(matches) != 3 {
		return 0, 0, ErrRangeFormat
	}

	start, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, ErrRangeStart
	}
	end, err = strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, ErrRangeEnd
	}

	if start < 0 || start > end {
		return 0, 0, ErrRangeNotSatisfiable
	}

	return start, end, nil
}
