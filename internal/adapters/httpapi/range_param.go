package httpapi

import (
	"encoding/json"
	"errors"

	"code/internal/application"
)

var (
	// ErrRangeFormat indicates the range value is not a JSON array [start,end].
	ErrRangeFormat = errors.New("invalid range, expected [start,end]")
	// ErrRangeInverted indicates end is lower than start.
	ErrRangeInverted = errors.New("invalid range, end must not be less than start")
)

// parseRangeParam parses a "range" query parameter value, a JSON array [start,end].
func parseRangeParam(rangeParam string) (application.Range, error) {
	var bounds []int64
	if err := json.Unmarshal([]byte(rangeParam), &bounds); err != nil || len(bounds) != 2 || bounds[0] < 0 {
		return application.Range{}, ErrRangeFormat
	}

	if bounds[0] > bounds[1] {
		return application.Range{}, ErrRangeInverted
	}

	return application.Range{First: bounds[0], Last: bounds[1]}, nil
}
