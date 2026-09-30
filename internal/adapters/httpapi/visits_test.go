package httpapi

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

func TestHandlerListVisitsRangeMapsRequest(t *testing.T) {
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(5, 9, application.Sort{})).
		Return(
			application.RangePage[application.Visit]{
				Items: make([]application.Visit, 5),
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

func TestHandlerListVisitsRangeSortMapsRequest(t *testing.T) {
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(
			0,
			4,
			application.Sort{Field: "created_at", Asc: false},
		)).
		Return(
			application.RangePage[application.Visit]{
				Items: make([]application.Visit, 5),
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
		items []application.Visit
	}{
		{name: "nil items", items: nil},
		{name: "empty items", items: []application.Visit{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visitService := NewMockUseCase(t)
			visitService.EXPECT().
				PageLinkVisits(mock.Anything, visitsQuery(0, 4, application.Sort{})).
				Return(
					application.RangePage[application.Visit]{
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
		Return(application.RangePage[application.Visit]{}, nil).
		Once()
	handler := newTestHandler(visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits */0", w.Header().Get("Content-Range"))
	assert.JSONEq(t, "[]", w.Body.String())
}

func TestHandlerListVisitsMapsResponseContract(t *testing.T) {
	visit := application.Visit{
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
		Return(application.RangePage[application.Visit]{Items: []application.Visit{visit}, Total: 1}, nil).
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

func TestHandlerListVisitsIgnoresForeignRangeUnit(t *testing.T) {
	visitService := NewMockUseCase(t)
	visitService.EXPECT().
		PageLinkVisits(mock.Anything, visitsQuery(0, firstPageLast, application.Sort{})).
		Return(application.RangePage[application.Visit]{
			Items: make([]application.Visit, 3),
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
