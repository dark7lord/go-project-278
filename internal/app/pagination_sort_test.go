package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func visitIP(i int) string {
	return fmt.Sprintf("10.0.0.%d", i+1)
}

func TestLinksRangeSortShortName(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	tx := setupTestTx(t, td)

	seedNames := []string{"banana", "apple", "cherry"}
	for _, name := range seedNames {
		_, err := tx.repo.CreateLink(ctx, "https://x.example/"+name, name)
		require.NoError(t, err)
	}

	tests := []struct {
		name string
		sort string
		want []int
	}{
		{name: "short_name ascending", sort: `["short_name","ASC"]`, want: []int{1, 0, 2}},
		{name: "short_name descending", sort: `["short_name","DESC"]`, want: []int{2, 0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr := "/api/links?" + url.Values{
				"range": {"[0,3]"},
				"sort":  {tt.sort},
			}.Encode()
			w := performRequest(t, tx.router, "GET", urlStr, "")

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "links 0-2/3", w.Header().Get("Content-Range"))

			var responses []linkResponse
			decode(t, w, &responses)
			require.Len(t, responses, len(tt.want))
			for i, response := range responses {
				assert.Equal(t, seedNames[tt.want[i]], response.ShortName)
			}
		})
	}
}

func TestVisitsRangeSortByIP(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.repo.CreateLink(ctx, link.OriginalURL, link.ShortName)
	require.NoError(t, err)

	seedIPs := []string{visitIP(1), visitIP(0), visitIP(2)}
	for _, ip := range seedIPs {
		_, err := tx.repo.CreateLinkVisit(ctx, created.ID, ip, "agent", nil, int32(http.StatusFound))
		require.NoError(t, err)
	}

	tests := []struct {
		name string
		sort string
		want []int
	}{
		{name: "ip ascending", sort: `["ip","ASC"]`, want: []int{1, 0, 2}},
		{name: "ip descending", sort: `["ip","DESC"]`, want: []int{2, 0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr := "/api/link_visits?" + url.Values{
				"range": {"[0,9]"},
				"sort":  {tt.sort},
			}.Encode()
			w := performRequest(t, tx.router, "GET", urlStr, "")

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "link_visits 0-2/3", w.Header().Get("Content-Range"))

			var visits []application.VisitView
			decode(t, w, &visits)
			require.Len(t, visits, len(tt.want))
			for i, visit := range visits {
				assert.Equal(t, seedIPs[tt.want[i]], visit.IP)
			}
		})
	}
}

func TestVisitsRangeSortRefererNullsLast(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.repo.CreateLink(ctx, link.OriginalURL, link.ShortName)
	require.NoError(t, err)

	refA := "https://a.example"
	refB := "https://b.example"
	for _, seed := range []struct {
		ip      string
		referer *string
	}{
		{ip: visitIP(0), referer: &refA},
		{ip: visitIP(1), referer: nil},
		{ip: visitIP(2), referer: &refB},
	} {
		_, err := tx.repo.CreateLinkVisit(ctx, created.ID, seed.ip, "agent", seed.referer, int32(http.StatusFound))
		require.NoError(t, err)
	}

	tests := []struct {
		name string
		sort string
		want []int
	}{
		{name: "referer ascending", sort: `["referer","ASC"]`, want: []int{0, 2, 1}},
		{name: "referer descending", sort: `["referer","DESC"]`, want: []int{2, 0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr := "/api/link_visits?" + url.Values{
				"range": {"[0,9]"},
				"sort":  {tt.sort},
			}.Encode()
			w := performRequest(t, tx.router, "GET", urlStr, "")

			assert.Equal(t, http.StatusOK, w.Code)

			var visits []application.VisitView
			decode(t, w, &visits)
			require.Len(t, visits, len(tt.want))
			for i, visit := range visits {
				assert.Equal(t, visitIP(tt.want[i]), visit.IP)
			}
		})
	}
}

func TestVisitsRangeSortCreatedAt(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.repo.CreateLink(ctx, link.OriginalURL, link.ShortName)
	require.NoError(t, err)

	// CURRENT_TIMESTAMP is constant within a transaction, so created_at is
	// backdated per visit to make the ordering deterministic.
	for i := range 3 {
		visit, err := tx.repo.CreateLinkVisit(ctx, created.ID, visitIP(i), "agent", nil, int32(http.StatusFound))
		require.NoError(t, err)

		_, err = tx.tx.Exec(
			ctx,
			"UPDATE link_visits SET created_at = now() - $1::int * INTERVAL '1 minute' WHERE id = $2",
			i,
			visit.ID,
		)
		require.NoError(t, err)
	}

	tests := []struct {
		name string
		sort string
		want []int
	}{
		{name: "created_at descending", sort: `["created_at","DESC"]`, want: []int{0, 1, 2}},
		{name: "created_at ascending", sort: `["created_at","ASC"]`, want: []int{2, 1, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr := "/api/link_visits?" + url.Values{
				"range": {"[0,9]"},
				"sort":  {tt.sort},
			}.Encode()
			w := performRequest(t, tx.router, "GET", urlStr, "")

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "link_visits 0-2/3", w.Header().Get("Content-Range"))

			var visits []application.VisitView
			decode(t, w, &visits)
			require.Len(t, visits, len(tt.want))
			for i, visit := range visits {
				assert.Equal(t, visitIP(tt.want[i]), visit.IP)
			}
		})
	}
}
