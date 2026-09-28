package httpadapter

import (
	"errors"
	"regexp"
	"strconv"

	"code/internal/application"
)

// rangeParamRe matches the range query parameter: [start,end].
var rangeParamRe = regexp.MustCompile(`^\s*\[\s*(\d+)\s*,\s*(\d+)\s*\]\s*$`)

var (
	// ErrRangeFormat indicates the range value does not match [start,end].
	ErrRangeFormat = errors.New("invalid range, expected [start,end]")
	// ErrRangeStart indicates the start value is not a number.
	ErrRangeStart = errors.New("invalid start value")
	// ErrRangeEnd indicates the end value is not a number.
	ErrRangeEnd = errors.New("invalid end value")
	// ErrRangeInverted indicates end is lower than start.
	ErrRangeInverted = errors.New("invalid range, end must not be less than start")
)

// parseRangeParam parses a "range" query parameter value: [start,end].
func parseRangeParam(rangeParam string) (application.Range, error) {
	matches := rangeParamRe.FindStringSubmatch(rangeParam)
	if len(matches) != 3 {
		return application.Range{}, ErrRangeFormat
	}

	start, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return application.Range{}, ErrRangeStart
	}
	end, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return application.Range{}, ErrRangeEnd
	}

	if start > end {
		return application.Range{}, ErrRangeInverted
	}

	return application.Range{First: start, Last: end}, nil
}
