package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/db"
)

func TestLinksPagination(t *testing.T) {
	// t.Skip("waiting for pagination implementation — handler/service/repo/queries need GetLinksRange and CountLinks")
	td := setupTestDB(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		rangeQuery string
		seedCount  int
		wantStatus int
		wantRange  string
		wantLen    int
	}{
		{
			name:       "first page",
			rangeQuery: "[0,4]",
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
			wantRange:  "links 0-4/15",
			wantLen:    5,
		},
		{
			name:       "middle page",
			rangeQuery: "[5,9]",
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
			wantRange:  "links 5-9/15",
			wantLen:    5,
		},
		{
			name:       "last partial page",
			rangeQuery: "[10,14]",
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
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
			name:       "range beyond total returns empty",
			rangeQuery: "[100,200]",
			seedCount:  5,
			wantStatus: http.StatusPartialContent,
			wantRange:  "links 100-200/5",
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

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", urlStr, nil)
			tx.router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantRange != "" {
				assert.Equal(t, tt.wantRange, w.Header().Get("Content-Range"))
			}

			if w.Code == http.StatusPartialContent || w.Code == http.StatusOK {
				var links []db.Link
				err := json.Unmarshal(w.Body.Bytes(), &links)
				require.NoError(t, err)
				assert.Len(t, links, tt.wantLen)
			}
		})
	}
}
