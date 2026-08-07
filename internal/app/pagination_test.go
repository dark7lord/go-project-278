package app

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/db"
)

const (
	range04     = "[0,4]"
	range09     = "[0,9]"
	links04     = "links 0-4/15"
	links59     = "links 5-9/15"
	links1014   = "links 10-14/15"
	links100200 = "links 100-200/5"
)

func TestLinksPagination(t *testing.T) {
	// t.Skip("waiting for pagination implementation — handler/service/repo/queries need GetLinksRange and CountLinks")
	td := setupTestDB(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		rangeQuery  string
		rangeHeader string
		seedCount   int
		wantStatus  int
		wantRange   string
		wantLen     int
	}{
		{
			name:       "first page",
			rangeQuery: range04,
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
			wantRange:  links04,
			wantLen:    5,
		},
		{
			name:       "middle page",
			rangeQuery: "[5,9]",
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
			wantRange:  links59,
			wantLen:    5,
		},
		{
			name:       "last partial page",
			rangeQuery: "[10,14]",
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
			wantRange:  links1014,
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
			name:       "range beyond total returns empty",
			rangeQuery: "[100,200]",
			seedCount:  5,
			wantStatus: http.StatusPartialContent,
			wantRange:  links100200,
			wantLen:    0,
		},
		{
			name:       "start > end",
			rangeQuery: "[10,5]",
			seedCount:  0,
			wantStatus: http.StatusRequestedRangeNotSatisfiable,
		},
		{
			name:       "bad format",
			rangeQuery: "invalid",
			seedCount:  0,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "range header applies without query param",
			rangeHeader: range04,
			seedCount:   15,
			wantStatus:  http.StatusPartialContent,
			wantRange:   links04,
			wantLen:     5,
		},
		{
			name:        "query param overrides range header",
			rangeQuery:  range04,
			rangeHeader: range09,
			seedCount:   15,
			wantStatus:  http.StatusPartialContent,
			wantRange:   links04,
			wantLen:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			for i := range tt.seedCount {
				l := linkFactory(i)
				_, err := tx.repo.CreateLink(ctx, l.OriginalURL, l.ShortName, l.ShortURL)
				require.NoError(t, err)
			}

			urlStr := "/api/links"
			if tt.rangeQuery != "" {
				urlStr += "?range=" + url.QueryEscape(tt.rangeQuery)
			}

			req, _ := http.NewRequest("GET", urlStr, nil)
			if tt.rangeHeader != "" {
				req.Header.Set("Range", tt.rangeHeader)
			}
			w := serve(t, tx.router, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantRange != "" {
				assert.Equal(t, tt.wantRange, w.Header().Get("Content-Range"))
			}

			if w.Code == http.StatusPartialContent || w.Code == http.StatusOK {
				var links []db.Link
				decode(t, w, &links)
				assert.Len(t, links, tt.wantLen)
			}
		})
	}
}
