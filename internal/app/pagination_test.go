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
			name:       "no range returns all",
			rangeQuery: "",
			seedCount:  15,
			wantStatus: http.StatusOK,
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
			name:        "range header applies without query param",
			rangeHeader: "[2,6]",
			seedCount:   15,
			start:       2,
			wantStatus:  http.StatusOK,
			wantRange:   "links 2-6/15",
			wantLen:     5,
		},
		{
			name:        "query param overrides range header",
			rangeQuery:  "[3,7]",
			rangeHeader: "[0,9]",
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

			if w.Code == http.StatusOK {
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
				if tt.rangeQuery == "" && tt.rangeHeader == "" {
					assert.Equal(t, seeds, links)

					return
				}
				assert.Equal(t, seeds[tt.start:tt.start+tt.wantLen], links)
			}
		})
	}
}
