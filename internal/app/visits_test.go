package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/db"
)

const visits04 = "visits 0-4/15"

func TestVisitsPagination(t *testing.T) {
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
			wantRange:  visits04,
			wantLen:    5,
		},
		{
			name:       "middle page",
			rangeQuery: "[5,9]",
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
			wantRange:  "visits 5-9/15",
			wantLen:    5,
		},
		{
			name:       "last partial page",
			rangeQuery: "[10,14]",
			seedCount:  15,
			wantStatus: http.StatusPartialContent,
			wantRange:  "visits 10-14/15",
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
			wantRange:  "visits 100-200/5",
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
			name:        "range header first page",
			rangeHeader: range04,
			seedCount:   15,
			wantStatus:  http.StatusPartialContent,
			wantRange:   visits04,
			wantLen:     5,
		},
		{
			name:        "query param overrides range header",
			rangeQuery:  range04,
			rangeHeader: range09,
			seedCount:   15,
			wantStatus:  http.StatusPartialContent,
			wantRange:   visits04,
			wantLen:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			link := linkFactory(0)
			created, err := tx.repo.CreateLink(ctx, link.OriginalURL, link.ShortName, link.ShortURL)
			require.NoError(t, err)

			for i := range tt.seedCount {
				ref := fmt.Sprintf("https://ref-%d.com", i)
				_, err := tx.repo.CreateLinkVisit(
					ctx, created.ID,
					fmt.Sprintf("10.0.0.%d", i),
					fmt.Sprintf("agent-%d", i),
					&ref,
					int32(http.StatusFound),
				)
				require.NoError(t, err)
			}

			urlStr := "/api/link_visits"
			if tt.rangeQuery != "" {
				urlStr += "?range=" + url.QueryEscape(tt.rangeQuery)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", urlStr, nil)
			if tt.rangeHeader != "" {
				req.Header.Set("Range", tt.rangeHeader)
			}
			tx.router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantRange != "" {
				assert.Equal(t, tt.wantRange, w.Header().Get("Content-Range"))
			}

			if w.Code == http.StatusPartialContent || w.Code == http.StatusOK {
				var visits []db.LinkVisit
				err := json.Unmarshal(w.Body.Bytes(), &visits)
				require.NoError(t, err)
				assert.Len(t, visits, tt.wantLen)
			}
		})
	}
}

func TestRedirectRecordsVisit(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.repo.CreateLink(ctx, link.OriginalURL, link.ShortName, link.ShortURL)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/r/"+created.ShortName, nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Referer", "https://example.com")
	tx.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	visits, err := tx.svc.ListLinkVisits(ctx)
	require.NoError(t, err)
	require.Len(t, visits, 1)

	assert.Equal(t, created.ID, visits[0].LinkID)
	assert.Equal(t, "192.0.2.1", visits[0].IP)
	assert.Equal(t, "test-agent", visits[0].UserAgent)
	require.NotNil(t, visits[0].Referer)
	assert.Equal(t, "https://example.com", *visits[0].Referer)
	assert.Equal(t, int32(http.StatusFound), visits[0].Status)
}
