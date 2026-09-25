package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func TestVisitsPagination(t *testing.T) {
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
			wantRange:  "link_visits 0-4/15",
			wantLen:    5,
		},
		{
			name:       "middle page",
			rangeQuery: "[5,9]",
			seedCount:  15,
			start:      5,
			wantStatus: http.StatusOK,
			wantRange:  "link_visits 5-9/15",
			wantLen:    5,
		},
		{
			name:       "last partial page",
			rangeQuery: "[10,14]",
			seedCount:  15,
			start:      10,
			wantStatus: http.StatusOK,
			wantRange:  "link_visits 10-14/15",
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
			wantRange:  "link_visits */5",
		},
		{
			name:       "start > end",
			rangeQuery: "[10,5]",
			seedCount:  0,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad format",
			rangeQuery: "invalid",
			seedCount:  0,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:        "range header first page",
			rangeHeader: "[2,6]",
			seedCount:   15,
			start:       2,
			wantStatus:  http.StatusOK,
			wantRange:   "link_visits 2-6/15",
			wantLen:     5,
		},
		{
			name:        "query param overrides range header",
			rangeQuery:  "[3,7]",
			rangeHeader: "[0,9]",
			seedCount:   15,
			start:       3,
			wantStatus:  http.StatusOK,
			wantRange:   "link_visits 3-7/15",
			wantLen:     5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			link := linkFactory(0)
			created, err := tx.repo.CreateLink(ctx, link.OriginalURL, link.ShortName)
			require.NoError(t, err)

			var seeds []application.VisitView
			for i := range tt.seedCount {
				ref := fmt.Sprintf("https://ref-%d.com", i)
				visit, err := tx.repo.CreateLinkVisit(
					ctx, created.ID,
					fmt.Sprintf("10.0.0.%d", i),
					fmt.Sprintf("agent-%d", i),
					&ref,
					int32(http.StatusFound),
				)
				require.NoError(t, err)
				seeds = append(seeds, visit)
			}

			urlStr := "/api/link_visits"
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
				var visits []application.VisitView
				decode(t, w, &visits)
				assert.Len(t, visits, tt.wantLen)

				if tt.wantLen > 0 {
					for i, visit := range visits {
						assert.Equal(t, seeds[tt.start+i].ID, visit.ID)
						assert.Equal(t, seeds[tt.start+i].IP, visit.IP)
					}
				}
			}
		})
	}
}

func TestRedirectRecordsVisit(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.repo.CreateLink(ctx, link.OriginalURL, link.ShortName)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/r/"+created.ShortName, nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Referer", "https://example.com")
	w := httptest.NewRecorder()
	tx.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	visits, err := tx.svc.ListLinkVisits(ctx)
	require.NoError(t, err)
	require.Len(t, visits, 1)

	assert.Equal(t, created.ID, visits[0].LinkID)
	assert.Equal(t, "192.0.2.1", visits[0].IP)
	assert.Equal(t, "test-agent", visits[0].UserAgent)
	assert.Equal(t, int32(http.StatusFound), visits[0].Status)

	stored, err := tx.queries.GetLinkVisits(ctx)
	require.NoError(t, err)
	require.Len(t, stored, 1)
	require.NotNil(t, stored[0].Referer)
	assert.Equal(t, "https://example.com", *stored[0].Referer)
}

func TestRedirectErrors(t *testing.T) {
	td := setupTestDB(t)

	t.Run("not found", func(t *testing.T) {
		tx := setupTestTx(t, td)

		w := performRequest(t, tx.router, "GET", "/r/nonexistent", "")

		assert.Equal(t, http.StatusNotFound, w.Code)
		assertErrorBody(t, w)
	})
}

func TestListVisitsEmpty(t *testing.T) {
	td := setupTestDB(t)

	tx := setupTestTx(t, td)

	w := performRequest(t, tx.router, "GET", "/api/link_visits", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}
