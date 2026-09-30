package httpadapter

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func TestHandlerListLinksRangeMapsRequest(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(5, 9, application.Sort{})).
		Return(application.RangePage[application.LinkView]{
			Items: make([]application.LinkView, 5),
			First: 5,
			Total: 10,
		}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/links?"+url.Values{"range": {"[5,9]"}}.Encode(),
		nil,
	)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 5-9/10", w.Header().Get("Content-Range"))
}

func TestHandlerListLinksRangeUnsatisfiable(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(100, 200, application.Sort{})).
		Return(application.RangePage[application.LinkView]{
			Items: []application.LinkView{},
			First: 100,
			Total: 5,
		}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/links?"+url.Values{"range": {"[100,200]"}}.Encode(),
		nil,
	)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "links */5", w.Header().Get("Content-Range"))
}

func TestHandlerListLinksRangeEmptyCollection(t *testing.T) {
	tests := []struct {
		name  string
		items []application.LinkView
	}{
		{name: "nil items", items: nil},
		{name: "empty items", items: []application.LinkView{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkService := NewMockUseCase(t)
			linkService.EXPECT().
				PageLinks(mock.Anything, linksQuery(0, 4, application.Sort{})).
				Return(
					application.RangePage[application.LinkView]{
						Items: tt.items,
						First: 0,
						Total: 0,
					},
					nil,
				).
				Once()
			handler := newTestHandler(linkService)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				"/links?"+url.Values{"range": {"[0,4]"}}.Encode(),
				nil,
			)
			newHandlerRouter(handler).ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "links */0", w.Header().Get("Content-Range"))
			assert.JSONEq(t, "[]", w.Body.String())
		})
	}
}

func TestHandlerListLinksRangeEmptyItems(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(5, 9, application.Sort{})).
		Return(application.RangePage[application.LinkView]{Items: []application.LinkView{}, First: 5, Total: 10}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+url.Values{"range": {"[5,9]"}}.Encode(), nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "links */10", w.Header().Get("Content-Range"))
}

func TestHandlerListLinksCutsRangeToMaximumPageSize(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(10, 10+firstPageLast, application.Sort{})).
		Return(application.RangePage[application.LinkView]{
			Items: make([]application.LinkView, maxPageSize),
			First: 10,
			Total: 5000,
		}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+url.Values{"range": {"[10,4999]"}}.Encode(), nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 10-1009/5000", w.Header().Get("Content-Range"))
}

func TestHandlerListLinksRejectsMalformedRange(t *testing.T) {
	for _, tc := range []struct {
		name     string
		rangeVal string
		wantBody string
	}{
		{
			name:     "bad format",
			rangeVal: "invalid",
			wantBody: `{"error": "invalid range, expected [start,end]"}`,
		},
		{
			name:     "start > end",
			rangeVal: "[10,5]",
			wantBody: `{"error": "invalid range, end must not be less than start"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			linkService := NewMockUseCase(t)
			handler := newTestHandler(linkService)

			w := httptest.NewRecorder()
			query := url.Values{"range": {tc.rangeVal}}.Encode()
			req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
			newHandlerRouter(handler).ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, tc.wantBody, w.Body.String())
			linkService.AssertNotCalled(t, "PageLinks", mock.Anything, mock.Anything)
		})
	}
}

func TestHandlerListLinksRangeSortMapsRequest(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(
			5,
			9,
			application.Sort{Field: "short_name", Asc: true},
		)).
		Return(application.RangePage[application.LinkView]{Items: make([]application.LinkView, 5), First: 5, Total: 10}, nil).
		Once()
	handler := newTestHandler(linkService)

	query := url.Values{
		"range": {"[5,9]"},
		"sort":  {`["short_name","ASC"]`},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 5-9/10", w.Header().Get("Content-Range"))
}

func TestHandlerListLinksSortWithoutRangeSortsFirstPage(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(
			0,
			firstPageLast,
			application.Sort{Field: "short_name", Asc: true},
		)).
		Return(application.RangePage[application.LinkView]{Items: make([]application.LinkView, 3), Total: 3}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	query := url.Values{
		"sort": {`["short_name","ASC"]`},
	}.Encode()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 0-2/3", w.Header().Get("Content-Range"))
}

func TestHandlerListLinksSortUnsupportedField(t *testing.T) {
	linkService := NewMockUseCase(t)
	handler := newTestHandler(linkService)

	query := url.Values{
		"range": {"[0,4]"},
		"sort":  {`["bogus","ASC"]`},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "unsupported sort field"}`, w.Body.String())
	linkService.AssertNotCalled(t, "PageLinks", mock.Anything, mock.Anything)
}

func TestHandlerListLinksSortBadFormat(t *testing.T) {
	linkService := NewMockUseCase(t)
	handler := newTestHandler(linkService)

	query := url.Values{
		"range": {"[0,4]"},
		"sort":  {"short_name"},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "invalid sort, expected [field,ASC|DESC]"}`, w.Body.String())
	linkService.AssertNotCalled(t, "PageLinks", mock.Anything, mock.Anything)
}

func TestHandlerListVisitsRangeMapsRequest(t *testing.T) {
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(5, 9, application.Sort{})).
		Return(
			application.RangePage[application.VisitView]{
				Items: make([]application.VisitView, 5),
				First: 5,
				Total: 10,
			},
			nil,
		).
		Once()
	handler := newTestHandler(visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits?"+url.Values{"range": {"[5,9]"}}.Encode(), nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 5-9/10", w.Header().Get("Content-Range"))
}

func TestHandlerListLinksEmptyCollection(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(0, firstPageLast, application.Sort{})).
		Return(application.RangePage[application.LinkView]{}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links */0", w.Header().Get("Content-Range"))
	assert.JSONEq(t, "[]", w.Body.String())
}

func TestHandlerListVisitsRangeSortMapsRequest(t *testing.T) {
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(
			0,
			4,
			application.Sort{Field: "created_at", Asc: false},
		)).
		Return(
			application.RangePage[application.VisitView]{
				Items: make([]application.VisitView, 5),
				First: 0,
				Total: 5,
			},
			nil,
		).
		Once()
	handler := newTestHandler(visitService)

	query := url.Values{
		"range": {"[0,4]"},
		"sort":  {`["created_at","DESC"]`},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-4/5", w.Header().Get("Content-Range"))
}

func TestHandlerListVisitsRangeEmptyCollection(t *testing.T) {
	tests := []struct {
		name  string
		items []application.VisitView
	}{
		{name: "nil items", items: nil},
		{name: "empty items", items: []application.VisitView{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visitService := NewMockUseCase(t)
			visitService.EXPECT().
				PageLinkVisits(mock.Anything, visitsQuery(0, 4, application.Sort{})).
				Return(
					application.RangePage[application.VisitView]{
						Items: tt.items,
						First: 0,
						Total: 0,
					},
					nil,
				).
				Once()
			handler := newTestHandler(visitService)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodGet,
				"/link_visits?"+url.Values{"range": {"[0,4]"}}.Encode(),
				nil,
			)
			newHandlerRouter(handler).ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "link_visits */0", w.Header().Get("Content-Range"))
			assert.JSONEq(t, "[]", w.Body.String())
		})
	}
}

func TestHandlerListVisitsEmptyCollection(t *testing.T) {
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(0, firstPageLast, application.Sort{})).
		Return(application.RangePage[application.VisitView]{}, nil).
		Once()
	handler := newTestHandler(visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits */0", w.Header().Get("Content-Range"))
	assert.JSONEq(t, "[]", w.Body.String())
}

func TestHandlerListLinksMapsResponseContract(t *testing.T) {
	link := application.LinkView{
		ID:          7,
		OriginalURL: testExampleURL,
		ShortName:   "target",
	}
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(0, firstPageLast, application.Sort{})).
		Return(application.RangePage[application.LinkView]{Items: []application.LinkView{link}, Total: 1}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var items []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	require.Len(t, items, 1)

	wantKeys := []string{
		"id",
		"original_url",
		"short_name",
		"short_url",
	}
	assert.ElementsMatch(t, wantKeys, slices.Collect(maps.Keys(items[0])))
}

func TestHandlerListVisitsMapsResponseContract(t *testing.T) {
	visit := application.VisitView{
		ID:        7,
		LinkID:    42,
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		IP:        "192.0.2.1",
		UserAgent: "test-agent",
		Status:    int32(http.StatusFound),
	}
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(0, firstPageLast, application.Sort{})).
		Return(application.RangePage[application.VisitView]{Items: []application.VisitView{visit}, Total: 1}, nil).
		Once()
	handler := newTestHandler(visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var items []map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	require.Len(t, items, 1)

	wantKeys := []string{
		"created_at",
		"id",
		"ip",
		"link_id",
		"reffer",
		"status",
		"user_agent",
	}
	assert.ElementsMatch(t, wantKeys, slices.Collect(maps.Keys(items[0])))
}

func TestHandlerListLinksRangeHeaderAnswersPartialContent(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, application.PageQuery{
			Range: application.Range{Suffix: true, Length: 2},
		}).
		Return(application.RangePage[application.LinkView]{
			Items: make([]application.LinkView, 2),
			First: 8,
			Total: 10,
		}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	req.Header.Set("Range", "links=-2")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusPartialContent, w.Code)
	assert.Equal(t, "links 8-9/10", w.Header().Get("Content-Range"))
	assert.Equal(t, "links", w.Header().Get("Accept-Ranges"))
}

func TestHandlerListVisitsIgnoresForeignRangeUnit(t *testing.T) {
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(0, firstPageLast, application.Sort{})).
		Return(application.RangePage[application.VisitView]{
			Items: make([]application.VisitView, 3),
			Total: 3,
		}, nil).
		Once()
	handler := newTestHandler(visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits", nil)
	req.Header.Set("Range", "links=0-1")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-2/3", w.Header().Get("Content-Range"))
	assert.Equal(t, "link_visits", w.Header().Get("Accept-Ranges"))
}
