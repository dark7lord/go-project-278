package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// maxPageSize caps a page; Content-Range tells the client what a cut range got.
const maxPageSize int64 = 1000

// Collection names double as the units of their Range headers.
const (
	linksUnit      = "links"
	linkVisitsUnit = "link_visits"
)

// requestRange picks the requested page, cut to maxPageSize: the strict range
// query wins, then a Range header it understands, otherwise "0-". fromHeader
// reports the header was used, which is answered with 206.
func requestRange(c *gin.Context, unit string) (r application.Range, fromHeader bool, err error) {
	if param := c.Query("range"); param != "" {
		r, err = parseRangeParam(param)
		if err != nil {
			return application.Range{}, false, err
		}

		return capRange(r), false, nil
	}

	if r, ok := parseRangeHeader(c.GetHeader("Range"), unit); ok {
		return capRange(r), true, nil
	}

	return capRange(application.Range{First: 0, Last: math.MaxInt64}), false, nil
}

// capRange cuts a range to at most maxPageSize items.
func capRange(r application.Range) application.Range {
	if r.Suffix {
		r.Length = min(r.Length, maxPageSize)
	} else if r.Last-r.First >= maxPageSize {
		r.Last = r.First + maxPageSize - 1
	}

	return r
}

// writeRangePage answers a page with its Content-Range: 200 [] for an empty
// collection, 416 for a range that misses, 206 for a Range header, else 200.
func writeRangePage[T any](
	c *gin.Context,
	unit string,
	fromHeader bool,
	page application.RangePage[T],
) {
	c.Header("Accept-Ranges", unit)

	if len(page.Items) == 0 {
		c.Header("Content-Range", fmt.Sprintf("%s */%d", unit, page.Total))
		if page.Total == 0 {
			c.JSON(http.StatusOK, page.Items)
			return
		}
		c.JSON(http.StatusRequestedRangeNotSatisfiable, errJSON("range not satisfiable"))

		return
	}

	status := http.StatusOK
	if fromHeader {
		status = http.StatusPartialContent
	}

	lastPos := page.First + int64(len(page.Items)) - 1
	c.Header("Content-Range", fmt.Sprintf("%s %d-%d/%d", unit, page.First, lastPos, page.Total))
	c.JSON(status, page.Items)
}

var (
	errRangeFormat   = errors.New("invalid range, expected [start,end]")
	errRangeInverted = errors.New("invalid range, end must not be less than start")
)

// parseRangeParam parses a "range" query parameter value, a JSON array [start,end].
func parseRangeParam(rangeParam string) (application.Range, error) {
	var bounds []int64
	if err := json.Unmarshal([]byte(rangeParam), &bounds); err != nil || len(bounds) != 2 || bounds[0] < 0 {
		return application.Range{}, errRangeFormat
	}

	if bounds[0] > bounds[1] {
		return application.Range{}, errRangeInverted
	}

	return application.Range{First: bounds[0], Last: bounds[1]}, nil
}

// rangeSpecRe matches a single RFC 9110 range-spec: first-last, first- or -suffix.
var rangeSpecRe = regexp.MustCompile(`^(\d*)-(\d*)$`)

// parseRangeHeader parses an RFC 9110 Range header: unit=a-b, unit=a- or
// unit=-n. Another unit, a bad value or several ranges (a comma never matches)
// report false, and the header is ignored.
func parseRangeHeader(header, unit string) (application.Range, bool) {
	gotUnit, spec, found := strings.Cut(header, "=")
	if !found || !strings.EqualFold(gotUnit, unit) {
		return application.Range{}, false
	}

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
