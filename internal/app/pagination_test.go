package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func TestLinksPagination(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		rangeQuery  string
		rangeHeader string
		seedCount   int
		start       int
		wantStatus  int
		wantRange   string
		wantLen     int
	}{
		{
			name:       "first page",
			rangeQuery: "[0,4]",
			seedCount:  15,
			start:      0,
			wantStatus: http.StatusOK,
			wantRange:  "links 0-4/15",
			wantLen:    5,
		},
		{
			name:       "middle page",
			rangeQuery: "[5,9]",
			seedCount:  15,
			start:      5,
			wantStatus: http.StatusOK,
			wantRange:  "links 5-9/15",
			wantLen:    5,
		},
		{
			name:       "last partial page",
			rangeQuery: "[10,14]",
			seedCount:  15,
			start:      10,
			wantStatus: http.StatusOK,
			wantRange:  "links 10-14/15",
			wantLen:    5,
		},
		{
			name:       "no range reads the first page",
			seedCount:  15,
			wantStatus: http.StatusOK,
			wantRange:  "links 0-14/15",
			wantLen:    15,
		},
		{
			name:       "range beyond total returns 416",
			rangeQuery: "[100,200]",
			seedCount:  5,
			wantStatus: http.StatusRequestedRangeNotSatisfiable,
			wantRange:  "links */5",
		},
		{
			name:        "closed range header answers 206",
			rangeHeader: "links=2-6",
			seedCount:   15,
			start:       2,
			wantStatus:  http.StatusPartialContent,
			wantRange:   "links 2-6/15",
			wantLen:     5,
		},
		{
			name:        "open range header reads to the end",
			rangeHeader: "links=10-",
			seedCount:   15,
			start:       10,
			wantStatus:  http.StatusPartialContent,
			wantRange:   "links 10-14/15",
			wantLen:     5,
		},
		{
			name:        "suffix range header reads the last items",
			rangeHeader: "links=-3",
			seedCount:   15,
			start:       12,
			wantStatus:  http.StatusPartialContent,
			wantRange:   "links 12-14/15",
			wantLen:     3,
		},
		{
			name:        "suffix longer than total reads everything",
			rangeHeader: "links=-50",
			seedCount:   15,
			wantStatus:  http.StatusPartialContent,
			wantRange:   "links 0-14/15",
			wantLen:     15,
		},
		{
			name:        "zero suffix returns 416",
			rangeHeader: "links=-0",
			seedCount:   15,
			wantStatus:  http.StatusRequestedRangeNotSatisfiable,
			wantRange:   "links */15",
		},
		{
			name:        "open range header past the end returns 416",
			rangeHeader: "links=20-",
			seedCount:   15,
			wantStatus:  http.StatusRequestedRangeNotSatisfiable,
			wantRange:   "links */15",
		},
		{
			name:        "range header on an empty collection answers 200 []",
			rangeHeader: "links=0-9",
			wantStatus:  http.StatusOK,
			wantRange:   "links */0",
		},
		{
			name:        "foreign range unit is ignored",
			rangeHeader: "bytes=0-1",
			seedCount:   15,
			wantStatus:  http.StatusOK,
			wantRange:   "links 0-14/15",
			wantLen:     15,
		},
		{
			name:        "query param overrides range header",
			rangeQuery:  "[3,7]",
			rangeHeader: "links=0-9",
			seedCount:   15,
			start:       3,
			wantStatus:  http.StatusOK,
			wantRange:   "links 3-7/15",
			wantLen:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			var seeds []application.LinkView
			for i := range tt.seedCount {
				l := linkFactory(i)
				created, err := tx.linkRepo.CreateLink(ctx, l.OriginalURL, l.ShortName)
				require.NoError(t, err)
				seeds = append(seeds, created)
			}

			urlStr := "/api/links"
			if tt.rangeQuery != "" {
				urlStr += "?" + url.Values{"range": {tt.rangeQuery}}.Encode()
			}

			req, _ := http.NewRequest("GET", urlStr, nil)
			if tt.rangeHeader != "" {
				req.Header.Set("Range", tt.rangeHeader)
			}
			w := httptest.NewRecorder()
			tx.router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantRange != "" {
				assert.Equal(t, tt.wantRange, w.Header().Get("Content-Range"))
			}

			if w.Code == http.StatusOK || w.Code == http.StatusPartialContent {
				var responses []linkResponse
				decode(t, w, &responses)
				links := make([]application.LinkView, len(responses))
				for index, response := range responses {
					links[index] = response.linkView()
				}
				assert.Len(t, links, tt.wantLen)

				if tt.wantLen == 0 {
					return
				}
				assert.Equal(t, seeds[tt.start:tt.start+tt.wantLen], links)
			}
		})
	}
}
