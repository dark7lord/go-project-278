package httpadapter

import (
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/application"
)

// maxPageSize bounds a page: a wider range is cut to it, and Content-Range
// tells the client which items it got.
const maxPageSize int64 = 1000

// Collection names double as the units of their Range headers.
const (
	linksUnit      = "links"
	linkVisitsUnit = "link_visits"
)

// requestRange picks the page a collection request asks for, cut to
// maxPageSize. The range query parameter (react-admin and the task send it) is
// strict and wins; the Range header follows RFC 9110 and is ignored when it is
// not understood; with neither, the request reads "0-", the first page.
// fromHeader reports an honoured Range header, which is answered with 206.
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

// contentRange builds a Content-Range value for a returned item range.
func contentRange(unit string, start, lastPos, total int64) string {
	return fmt.Sprintf("%s %d-%d/%d", unit, start, lastPos, total)
}

// unsatisfiedRange builds a Content-Range value for an unsatisfied range.
func unsatisfiedRange(unit string, total int64) string {
	return fmt.Sprintf("%s */%d", unit, total)
}

// writeRangePage writes a page with its Content-Range: 416 when the range
// misses a non-empty collection, 206 for an honoured Range header, 200
// otherwise. An empty collection answers 200 [] whatever the range, so a
// client sees an empty list rather than an error.
func writeRangePage[T any](
	c *gin.Context,
	unit string,
	fromHeader bool,
	page application.RangePage[T],
) {
	c.Header("Accept-Ranges", unit)

	if page.Total == 0 {
		c.Header("Content-Range", unsatisfiedRange(unit, page.Total))
		c.JSON(http.StatusOK, page.Items)
		return
	}

	if len(page.Items) == 0 {
		c.Header("Content-Range", unsatisfiedRange(unit, page.Total))
		c.JSON(http.StatusRequestedRangeNotSatisfiable, errJSON("range not satisfiable"))
		return
	}

	status := http.StatusOK
	if fromHeader {
		status = http.StatusPartialContent
	}

	lastPos := page.First + int64(len(page.Items)) - 1
	c.Header("Content-Range", contentRange(unit, page.First, lastPos, page.Total))
	c.JSON(status, page.Items)
}
