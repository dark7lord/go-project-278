package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

const (
	testBaseURL    = "https://short.example"
	testExampleURL = "https://example.com"
)

func newTestHandler(
	linkService application.LinkUseCase,
	visitService application.VisitUseCase,
) *Handler {
	return NewHandler(linkService, visitService, testBaseURL)
}

func newHandlerRouter(handler *Handler) *gin.Engine {
	router := gin.New()
	handler.RegisterAPIRoutes(router.Group(""))
	router.GET("/r/:code", handler.Redirect)

	return router
}

func TestHandlerCreateLinkValidationDoesNotCallUseCase(t *testing.T) {
	linkService := &mockLinkUseCase{}
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/links", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	linkService.AssertNotCalled(t, "CreateLink", mock.Anything, mock.Anything)
}

func TestHandlerIDValidationDoesNotCallUseCase(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/links/0"},
		{method: http.MethodGet, path: "/links/-1"},
		{method: http.MethodGet, path: "/links/abc"},
		{method: http.MethodPut, path: "/links/00"},
		{method: http.MethodPut, path: "/links/-2"},
		{method: http.MethodPut, path: "/links/xyz"},
		{method: http.MethodDelete, path: "/links/000"},
		{method: http.MethodDelete, path: "/links/-3"},
		{method: http.MethodDelete, path: "/links/12a"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			linkService := &mockLinkUseCase{}
			handler := newTestHandler(linkService, &mockVisitUseCase{})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			newHandlerRouter(handler).ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, `{"error": "invalid id"}`, w.Body.String())
			linkService.AssertNotCalled(t, "GetLinkByID", mock.Anything, mock.Anything)
			linkService.AssertNotCalled(t, "UpdateLink", mock.Anything, mock.Anything, mock.Anything)
			linkService.AssertNotCalled(t, "DeleteLink", mock.Anything, mock.Anything)
		})
	}
}

func TestHandlerLinkBindErrorsMapToUnprocessable(t *testing.T) {
	tooLongName := strings.Repeat("a", 33)

	for _, tc := range []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "create / absent original_url",
			method:     http.MethodPost,
			path:       "/links",
			body:       `{"short_name": "ok-link"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errors": {"original_url": "field is required"}}`,
		},
		{
			name:       "create / empty original_url",
			method:     http.MethodPost,
			path:       "/links",
			body:       `{"original_url": "", "short_name": "ok-link"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errors": {"original_url": "field is required"}}`,
		},
		{
			name:       "create / short name too short",
			method:     http.MethodPost,
			path:       "/links",
			body:       `{"original_url": "https://short.example", "short_name": "x"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errors": {"short_name": "must be at least 3 characters"}}`,
		},
		{
			name:       "create / short name too long",
			method:     http.MethodPost,
			path:       "/links",
			body:       `{"original_url": "https://long.example", "short_name": "` + tooLongName + `"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errors": {"short_name": "must be at most 32 characters"}}`,
		},
		{
			name:       "update / absent original_url",
			method:     http.MethodPut,
			path:       "/links/1",
			body:       `{"short_name": "ok-link"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errors": {"original_url": "field is required"}}`,
		},
		{
			name:       "update / short name too short",
			method:     http.MethodPut,
			path:       "/links/1",
			body:       `{"original_url": "https://short.example", "short_name": "x"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantBody:   `{"errors": {"short_name": "must be at least 3 characters"}}`,
		},
		{
			name:       "create / malformed json",
			method:     http.MethodPost,
			path:       "/links",
			body:       `{broken`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error": "invalid request"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			linkService := &mockLinkUseCase{}
			handler := newTestHandler(linkService, &mockVisitUseCase{})

			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			newHandlerRouter(handler).ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)
			assert.JSONEq(t, tc.wantBody, w.Body.String())
			linkService.AssertNotCalled(t, "CreateLink", mock.Anything, mock.Anything)
			linkService.AssertNotCalled(t, "UpdateLink", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestHandlerCreateLinkFieldErrorMapsToUnprocessable(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("CreateLink", mock.Anything, application.CreateLinkCommand{
			OriginalURL: "ftp://example.com",
			ShortName:   "ok-link",
		}).
		Return(application.LinkView{}, &application.FieldError{
			Field: "original_url",
			Err:   errors.New("unsupported scheme"),
		}).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(`{"original_url": "ftp://example.com", "short_name": "ok-link"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.JSONEq(t, `{"errors": {"original_url": "unsupported scheme"}}`, w.Body.String())
	linkService.AssertExpectations(t)
}

func TestHandlerUpdateLinkFieldErrorMapsToUnprocessable(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("UpdateLink", mock.Anything, int64(1), application.UpdateLinkCommand{
			OriginalURL: "https://example.com",
			ShortName:   "bad/name",
		}).
		Return(application.LinkView{}, &application.FieldError{
			Field: "short_name",
			Err:   errors.New("invalid short code"),
		}).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	req := httptest.NewRequest(
		http.MethodPut,
		"/links/1",
		strings.NewReader(`{"original_url": "https://example.com", "short_name": "bad/name"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.JSONEq(t, `{"errors": {"short_name": "invalid short code"}}`, w.Body.String())
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksRangeMapsRequest(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{Start: 5, End: 9}).
		Return(application.RangePage[application.LinkView]{
			Items: make([]application.LinkView, 5),
			Start: 5,
			Total: 10,
		}, nil).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/links?"+url.Values{"range": {"[5,9]"}}.Encode(),
		nil,
	)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 5-9/10", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksRangeUnsatisfiable(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{Start: 100, End: 200}).
		Return(application.RangePage[application.LinkView]{
			Items: []application.LinkView{},
			Start: 100,
			Total: 5,
		}, nil).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/links?"+url.Values{"range": {"[100,200]"}}.Encode(),
		nil,
	)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "links */5", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksRangeEmptyCollection(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{Start: 0, End: 4}).
		Return(
			application.RangePage[application.LinkView]{
				Items: []application.LinkView{},
				Start: 0,
				Total: 0,
			},
			nil,
		).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodGet,
		"/links?"+url.Values{"range": {"[0,4]"}}.Encode(),
		nil,
	)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links */0", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksRangeEmptyItems(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{Start: 5, End: 9}).
		Return(application.RangePage[application.LinkView]{Items: []application.LinkView{}, Start: 5, Total: 10}, nil).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+url.Values{"range": {"[5,9]"}}.Encode(), nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "links */10", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksRejectsRangeOverMaximumPageSize(t *testing.T) {
	linkService := &mockLinkUseCase{}
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+url.Values{"range": {"[0,1000]"}}.Encode(), nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "range exceeds maximum page size of 1000"}`, w.Body.String())
	linkService.AssertNotCalled(t, "ListLinksRange", mock.Anything, mock.Anything)
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
			linkService := &mockLinkUseCase{}
			handler := newTestHandler(linkService, &mockVisitUseCase{})

			w := httptest.NewRecorder()
			query := url.Values{"range": {tc.rangeVal}}.Encode()
			req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
			newHandlerRouter(handler).ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, tc.wantBody, w.Body.String())
			linkService.AssertNotCalled(t, "ListLinks", mock.Anything)
			linkService.AssertNotCalled(t, "ListLinksRange", mock.Anything, mock.Anything)
		})
	}
}

func TestHandlerListLinksRangeSortMapsRequest(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{
			Start: 5,
			End:   9,
			Sort:  &application.Sort{Field: application.SortFieldShortName, Asc: true},
		}).
		Return(application.RangePage[application.LinkView]{Items: make([]application.LinkView, 5), Start: 5, Total: 10}, nil).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	query := url.Values{
		"range": {"[5,9]"},
		"sort":  {`["short_name","ASC"]`},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 5-9/10", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksSortWithoutRange(t *testing.T) {
	linkService := &mockLinkUseCase{}
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	query := url.Values{
		"sort": {`["short_name","ASC"]`},
	}.Encode()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "sort requires a range"}`, w.Body.String())
	linkService.AssertNotCalled(t, "ListLinksRange", mock.Anything, mock.Anything)
}

func TestHandlerListLinksSortUnsupportedField(t *testing.T) {
	linkService := &mockLinkUseCase{}
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	query := url.Values{
		"range": {"[0,4]"},
		"sort":  {`["bogus","ASC"]`},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "unsupported sort field"}`, w.Body.String())
	linkService.AssertNotCalled(t, "ListLinksRange", mock.Anything, mock.Anything)
}

func TestHandlerListLinksSortBadFormat(t *testing.T) {
	linkService := &mockLinkUseCase{}
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	query := url.Values{
		"range": {"[0,4]"},
		"sort":  {"short_name"},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "invalid sort, expected [field,ASC|DESC]"}`, w.Body.String())
	linkService.AssertNotCalled(t, "ListLinksRange", mock.Anything, mock.Anything)
}

func TestHandlerRedirectMapsVisitMetadata(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("Redirect", mock.Anything, application.RedirectCommand{
			ShortName: "target",
			VisitMeta: application.VisitMeta{
				IP:        "192.0.2.1",
				UserAgent: "test-agent",
			},
		}).
		Return(application.LinkView{OriginalURL: testExampleURL}, nil).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/r/target", nil)
	req.Header.Set("User-Agent", "test-agent")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, testExampleURL, w.Header().Get("Location"))
	linkService.AssertExpectations(t)
}

func TestHandlerListVisitsRangeMapsRequest(t *testing.T) {
	visitService := &mockVisitUseCase{}
	visitService.
		On("ListLinkVisitsRange", mock.Anything, application.ListLinkVisitsQuery{Start: 0, End: 4}).
		Return(
			application.RangePage[application.VisitView]{
				Items: make([]application.VisitView, 5),
				Start: 0,
				Total: 5,
			},
			nil,
		).
		Once()
	handler := newTestHandler(&mockLinkUseCase{}, visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits?"+url.Values{"range": {"[0,4]"}}.Encode(), nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-4/5", w.Header().Get("Content-Range"))
	visitService.AssertExpectations(t)
}

func TestHandlerListVisitsRangeSortMapsRequest(t *testing.T) {
	visitService := &mockVisitUseCase{}
	visitService.
		On("ListLinkVisitsRange", mock.Anything, application.ListLinkVisitsQuery{
			Start: 0,
			End:   4,
			Sort:  &application.Sort{Field: application.SortFieldCreatedAt, Asc: false},
		}).
		Return(
			application.RangePage[application.VisitView]{
				Items: make([]application.VisitView, 5),
				Start: 0,
				Total: 5,
			},
			nil,
		).
		Once()
	handler := newTestHandler(&mockLinkUseCase{}, visitService)

	query := url.Values{
		"range": {"[0,4]"},
		"sort":  {`["created_at","DESC"]`},
	}.Encode()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits?"+query, nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-4/5", w.Header().Get("Content-Range"))
	visitService.AssertExpectations(t)
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
	visitService := &mockVisitUseCase{}
	visitService.
		On("ListLinkVisits", mock.Anything).
		Return([]application.VisitView{visit}, nil).
		Once()
	handler := newTestHandler(&mockLinkUseCase{}, visitService)

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
		"status",
		"user_agent",
	}
	assert.ElementsMatch(t, wantKeys, slices.Collect(maps.Keys(items[0])))
	visitService.AssertExpectations(t)
}

func TestHandlerListVisitsRejectsRangeOverMaximumPageSize(t *testing.T) {
	visitService := &mockVisitUseCase{}
	handler := newTestHandler(&mockLinkUseCase{}, visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/link_visits?"+url.Values{"range": {"[0,1000]"}}.Encode(), nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "range exceeds maximum page size of 1000"}`, w.Body.String())
	visitService.AssertNotCalled(t, "ListLinkVisitsRange", mock.Anything, mock.Anything)
}

func TestHandlerGetLinkRejectsInvalidID(t *testing.T) {
	handler := newTestHandler(&mockLinkUseCase{}, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links/not-an-id", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, errInvalidID, body["error"])
}

func TestHandlerCreateLinkHidesInternalError(t *testing.T) {
	rawErr := errors.New("postgres: dial tcp 127.0.0.1:5432: connection refused")
	linkService := &mockLinkUseCase{}
	linkService.
		On("CreateLink", mock.Anything, application.CreateLinkCommand{
			OriginalURL: "https://boom.example",
			ShortName:   "boomlink",
		}).
		Return(application.LinkView{}, rawErr).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(`{"original_url": "https://boom.example", "short_name": "boomlink"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error": "internal error"}`, w.Body.String())
	assert.NotContains(t, w.Body.String(), "connection refused")
	linkService.AssertExpectations(t)
}

func TestHandlerContextDeadlineMapsTo503(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinks", mock.Anything).
		Return([]application.LinkView{}, context.DeadlineExceeded).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error": "request timeout"}`, w.Body.String())
	linkService.AssertExpectations(t)
}

func TestHandlerUpdateLinkHidesInternalError(t *testing.T) {
	rawErr := errors.New("postgres: dial tcp 127.0.0.1:5432: connection refused")
	linkService := &mockLinkUseCase{}
	linkService.
		On("UpdateLink", mock.Anything, int64(1), application.UpdateLinkCommand{
			OriginalURL: "https://boom.example",
			ShortName:   "boomlink",
		}).
		Return(application.LinkView{}, rawErr).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPut,
		"/links/1",
		strings.NewReader(`{"original_url": "https://boom.example", "short_name": "boomlink"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error": "internal error"}`, w.Body.String())
	assert.NotContains(t, w.Body.String(), "connection refused")
	linkService.AssertExpectations(t)
}

func TestHandlerCreateLinkBuildsShortURL(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("CreateLink", mock.Anything, application.CreateLinkCommand{
			OriginalURL: testExampleURL,
			ShortName:   "example",
		}).
		Return(application.LinkView{
			ID:          7,
			OriginalURL: testExampleURL,
			ShortName:   "example",
		}, nil).
		Once()
	handler := newTestHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(`{"original_url": "https://example.com", "short_name": "example"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.JSONEq(t, `{
		"id": 7,
		"original_url": "https://example.com",
		"short_name": "example",
		"short_url": "https://short.example/r/example"
	}`, w.Body.String())
	linkService.AssertExpectations(t)
}
