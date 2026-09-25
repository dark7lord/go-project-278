package httpadapter

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

var rangeRe = regexp.MustCompile(`^\s*\[\s*(\d+)\s*,\s*(\d+)\s*\]\s*$`)

const maxPageSize int64 = 1000

var (
	// ErrRangeFormat indicates the range value does not match [start,end].
	ErrRangeFormat = errors.New("invalid range, expected [start,end]")
	// ErrRangeStart indicates the start value is not a number.
	ErrRangeStart = errors.New("invalid start value")
	// ErrRangeEnd indicates the end value is not a number.
	ErrRangeEnd = errors.New("invalid end value")
	// ErrRangeInverted indicates end is lower than start.
	ErrRangeInverted = errors.New("invalid range, end must not be less than start")
	// ErrRangeTooLarge indicates a range requests more than maxPageSize items.
	ErrRangeTooLarge = errors.New("range exceeds maximum page size of 1000")
)

// parseRangeParam parses a "range" query parameter value into start and end.
func parseRangeParam(rangeParam string) (start, end int64, err error) {
	matches := rangeRe.FindStringSubmatch(rangeParam)
	if len(matches) != 3 {
		return 0, 0, ErrRangeFormat
	}

	start, err = strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, 0, ErrRangeStart
	}
	end, err = strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return 0, 0, ErrRangeEnd
	}

	if start > end {
		return 0, 0, ErrRangeInverted
	}
	if end-start >= maxPageSize {
		return 0, 0, ErrRangeTooLarge
	}

	return start, end, nil
}

// contentRange builds a Content-Range value for a returned item range.
func contentRange(collection string, start, lastPos, total int64) string {
	return fmt.Sprintf("%s %d-%d/%d", collection, start, lastPos, total)
}

// unsatisfiedRange builds a Content-Range value for an unsatisfied range.
func unsatisfiedRange(collection string, total int64) string {
	return fmt.Sprintf("%s */%d", collection, total)
}

// writeRangePage resolves the range status against the page total and writes
// the Content-Range header and response body for a paginated page.
func writeRangePage[T any](c *gin.Context, collection string, page application.RangePage[T]) {
	if page.Total == 0 {
		c.Header("Content-Range", unsatisfiedRange(collection, page.Total))
		c.JSON(http.StatusOK, page.Items)
		return
	}

	if page.Start >= page.Total || len(page.Items) == 0 {
		c.Header("Content-Range", unsatisfiedRange(collection, page.Total))
		c.JSON(http.StatusRequestedRangeNotSatisfiable, errJSON("range not satisfiable"))
		return
	}

	c.Header("Content-Range", contentRange(collection, page.Start, page.Start+int64(len(page.Items))-1, page.Total))
	c.JSON(http.StatusOK, page.Items)
}
