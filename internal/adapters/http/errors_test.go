package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func TestHandlerCreateLinkValidationDoesNotCallUseCase(t *testing.T) {
	linkService := NewMockUseCase(t)
	handler := newTestHandler(linkService)

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
			linkService := NewMockUseCase(t)
			handler := newTestHandler(linkService)

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
		{
			name:       "update / unknown field",
			method:     http.MethodPut,
			path:       "/links/1",
			body:       `{"original_url": "https://strict-update.com", "short_name": "ok-link", "short_url": "boom"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error": "invalid request"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			linkService := NewMockUseCase(t)
			handler := newTestHandler(linkService)

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
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		CreateLink(mock.Anything, application.LinkInput{
			OriginalURL: "ftp://example.com",
			ShortName:   "ok-link",
		}).
		Return(application.LinkView{}, &application.FieldError{
			Field: "original_url",
			Err:   errors.New("unsupported scheme"),
		}).
		Once()
	handler := newTestHandler(linkService)

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
}

func TestHandlerUpdateLinkFieldErrorMapsToUnprocessable(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		UpdateLink(mock.Anything, int64(1), application.LinkInput{
			OriginalURL: "https://example.com",
			ShortName:   "bad/name",
		}).
		Return(application.LinkView{}, &application.FieldError{
			Field: "short_name",
			Err:   errors.New("invalid short code"),
		}).
		Once()
	handler := newTestHandler(linkService)

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
}

func TestHandlerGetLinkRejectsInvalidID(t *testing.T) {
	handler := newTestHandler(NewMockUseCase(t))

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
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		CreateLink(mock.Anything, application.LinkInput{
			OriginalURL: "https://boom.example",
			ShortName:   "boomlink",
		}).
		Return(application.LinkView{}, rawErr).
		Once()
	handler := newTestHandler(linkService)

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
}

func TestHandlerContextDeadlineMapsTo503(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		PageLinks(mock.Anything, linksQuery(0, firstPageLast, application.Sort{})).
		Return(application.RangePage[application.LinkView]{}, context.DeadlineExceeded).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"error": "request timeout"}`, w.Body.String())
}

func TestHandlerUpdateLinkHidesInternalError(t *testing.T) {
	rawErr := errors.New("postgres: dial tcp 127.0.0.1:5432: connection refused")
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		UpdateLink(mock.Anything, int64(1), application.LinkInput{
			OriginalURL: "https://boom.example",
			ShortName:   "boomlink",
		}).
		Return(application.LinkView{}, rawErr).
		Once()
	handler := newTestHandler(linkService)

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
}

func TestValidationErrorsNameFieldsByJSONTag(t *testing.T) {
	var body struct {
		Target string `json:"target_url" binding:"required"`
	}

	err := binding.Validator.ValidateStruct(&body)

	var validationErrors validator.ValidationErrors
	require.ErrorAs(t, err, &validationErrors)
	assert.Equal(t, "target_url", validationErrors[0].Field(), "the json tag, not the Go name")
}
