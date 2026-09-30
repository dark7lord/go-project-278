package app

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"slices"
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
	ctx := t.Context()

	tx := setupTestTx(t, td)

	seedNames := []string{"banana", "apple", "cherry"}
	for _, name := range seedNames {
		_, err := tx.linkRepo.CreateLink(ctx, "https://x.example/"+name, name)
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
	ctx := t.Context()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.linkRepo.CreateLink(ctx, link.OriginalURL, link.ShortName)
	require.NoError(t, err)

	seedIPs := []string{visitIP(1), visitIP(0), visitIP(2)}
	for _, ip := range seedIPs {
		_, err := tx.visitRepo.CreateLinkVisit(ctx, created.ID, foundVisit(ip))
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

			var visits []visitResponse
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
	ctx := t.Context()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.linkRepo.CreateLink(ctx, link.OriginalURL, link.ShortName)
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
		visit := foundVisit(seed.ip)
		visit.Referer = seed.referer
		_, err := tx.visitRepo.CreateLinkVisit(ctx, created.ID, visit)
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

			var visits []visitResponse
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
	ctx := t.Context()

	tx := setupTestTx(t, td)

	link := linkFactory(0)
	created, err := tx.linkRepo.CreateLink(ctx, link.OriginalURL, link.ShortName)
	require.NoError(t, err)

	// CURRENT_TIMESTAMP is constant within a transaction, so created_at is
	// backdated per visit to make the ordering deterministic.
	for i := range 3 {
		visit, err := tx.visitRepo.CreateLinkVisit(ctx, created.ID, foundVisit(visitIP(i)))
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

			var visits []visitResponse
			decode(t, w, &visits)
			require.Len(t, visits, len(tt.want))
			for i, visit := range visits {
				assert.Equal(t, visitIP(tt.want[i]), visit.IP)
			}
		})
	}
}

// sortedPage reads one page of a collection in the given sort order.
func sortedPage[T any](t *testing.T, tx *testDB, path, sort string) []T {
	t.Helper()

	query := url.Values{"range": {"[0,9]"}, "sort": {sort}}.Encode()
	w := performRequest(t, tx.router, http.MethodGet, path+"?"+query, "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var items []T
	decode(t, w, &items)

	return items
}

// assertSortsBothWays checks that field orders a page ascending and descending.
func assertSortsBothWays[T any](t *testing.T, tx *testDB, path, field string, compare func(a, b T) int) {
	t.Helper()

	asc := sortedPage[T](t, tx, path, `["`+field+`","ASC"]`)
	require.Len(t, asc, 3)
	assert.True(t, slices.IsSortedFunc(asc, compare), "ASC by %s: %+v", field, asc)

	desc := sortedPage[T](t, tx, path, `["`+field+`","DESC"]`)
	descending := func(a, b T) int { return compare(b, a) }
	assert.True(t, slices.IsSortedFunc(desc, descending), "DESC by %s: %+v", field, desc)
}

// TestLinksSortEveryField covers every links sort field the page query knows,
// on seeds whose field order differs from their insertion order.
func TestLinksSortEveryField(t *testing.T) {
	td := setupTestDB(t)
	tx := setupTestTx(t, td)

	for _, seed := range []struct{ url, name string }{
		{url: "https://c.example", name: "banana"},
		{url: "https://a.example", name: "cherry"},
		{url: "https://b.example", name: "apple"},
	} {
		_, err := tx.linkRepo.CreateLink(t.Context(), seed.url, seed.name)
		require.NoError(t, err)
	}

	for _, tt := range []struct {
		field   string
		compare func(a, b linkResponse) int
	}{
		{"id", func(a, b linkResponse) int { return cmp.Compare(a.ID, b.ID) }},
		{"short_name", func(a, b linkResponse) int { return cmp.Compare(a.ShortName, b.ShortName) }},
		{"short_url", func(a, b linkResponse) int { return cmp.Compare(a.ShortURL, b.ShortURL) }},
		{"original_url", func(a, b linkResponse) int { return cmp.Compare(a.OriginalURL, b.OriginalURL) }},
	} {
		t.Run(tt.field, func(t *testing.T) {
			assertSortsBothWays(t, tx, "/api/links", tt.field, tt.compare)
		})
	}
}

// TestVisitsSortEveryField does the same for the visits fields not covered by
// the dedicated ip, referer and created_at tests above.
func TestVisitsSortEveryField(t *testing.T) {
	td := setupTestDB(t)
	tx := setupTestTx(t, td)

	linkIDs := make([]int64, 3)
	for i := range linkIDs {
		l := linkFactory(i)
		created, err := tx.linkRepo.CreateLink(t.Context(), l.OriginalURL, l.ShortName)
		require.NoError(t, err)
		linkIDs[i] = created.ID
	}

	for _, seed := range []struct {
		link   int
		agent  string
		status int32
	}{
		{link: 2, agent: "b-agent", status: 302},
		{link: 0, agent: "c-agent", status: 301},
		{link: 1, agent: "a-agent", status: 307},
	} {
		_, err := tx.visitRepo.CreateLinkVisit(t.Context(), linkIDs[seed.link], application.Visit{
			IP:        visitIP(0),
			UserAgent: seed.agent,
			Status:    seed.status,
		})
		require.NoError(t, err)
	}

	for _, tt := range []struct {
		field   string
		compare func(a, b visitResponse) int
	}{
		{"id", func(a, b visitResponse) int { return cmp.Compare(a.ID, b.ID) }},
		{"link_id", func(a, b visitResponse) int { return cmp.Compare(a.LinkID, b.LinkID) }},
		{"user_agent", func(a, b visitResponse) int { return cmp.Compare(a.UserAgent, b.UserAgent) }},
		{"status", func(a, b visitResponse) int { return cmp.Compare(a.Status, b.Status) }},
	} {
		t.Run(tt.field, func(t *testing.T) {
			assertSortsBothWays(t, tx, "/api/link_visits", tt.field, tt.compare)
		})
	}
}
