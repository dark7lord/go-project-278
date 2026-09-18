package app

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testFieldOriginalURL = "original_url"
	testFieldShortName   = "short_name"
)

func TestUpdateLinkValidation(t *testing.T) {
	td := setupTestDB(t)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantField  string
	}{
		{
			name:       "missing original_url",
			body:       `{"original_url": "", "short_name": "ok-link"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldOriginalURL,
		},
		{
			name:       "original_url absent",
			body:       `{"short_name": "ok-link"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldOriginalURL,
		},
		{
			name:       "short name too short",
			body:       `{"original_url": "https://short.com", "short_name": "x"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
		{
			name:       "short name too long",
			body:       `{"original_url": "https://long.com", "short_name": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
		{
			name:       "short name cannot contain slash",
			body:       `{"original_url": "https://update-slash.com", "short_name": "ab/cd"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)
			ctx := context.Background()
			created, err := tx.repo.CreateLink(ctx, "https://update-validation.com", "update-validation", "http://localhost:8080/update-validation")
			require.NoError(t, err)

			w := performRequest(t, tx.router, "PUT", fmt.Sprintf("/api/links/%d", created.ID), tt.body)

			assert.Equal(t, tt.wantStatus, w.Code)
			assertFieldErrors(t, w, tt.wantField)
		})
	}
}

func TestCreateLinkValidation(t *testing.T) {
	td := setupTestDB(t)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantField  string
		wantError  string
	}{
		{
			name:       "missing original_url",
			body:       `{"original_url": "", "short_name": "ok-link"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldOriginalURL,
		},
		{
			name:       "unsupported scheme",
			body:       `{"original_url": "ftp://example.com", "short_name": "ok-link"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldOriginalURL,
		},
		{
			name:       "short name too short",
			body:       `{"original_url": "https://short.com", "short_name": "x"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
		{
			name:       "short name too long",
			body:       `{"original_url": "https://long.com", "short_name": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
		{
			name:       "short name cannot contain slash",
			body:       `{"original_url": "https://slash.com", "short_name": "ab/cd"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
		{
			name:       "short name cannot contain query characters",
			body:       `{"original_url": "https://query.com", "short_name": "abc?x=1"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
		{
			name:       "short name cannot contain fragment",
			body:       `{"original_url": "https://frag.com", "short_name": "abc#frag"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  testFieldShortName,
		},
		{
			name:       "invalid json",
			body:       `{broken`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request",
		},
		{
			name:       "unknown field",
			body:       `{"original_url": "https://strict.com", "short_name": "ok-link", "short_nam": "zzz"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			w := performRequest(t, tx.router, "POST", "/api/links", tt.body)

			if tt.wantField != "" {
				assertFieldErrors(t, w, tt.wantField)
			} else {
				assert.JSONEq(t, `{"error": "`+tt.wantError+`"}`, w.Body.String())
			}
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestUpdateLinkRejectsUnknownField(t *testing.T) {
	td := setupTestDB(t)
	tx := setupTestTx(t, td)
	ctx := context.Background()
	created, err := tx.repo.CreateLink(ctx, "https://strict-update.com", "strict-update", "http://localhost:8080/strict-update")
	require.NoError(t, err)

	w := performRequest(t, tx.router, "PUT", fmt.Sprintf("/api/links/%d", created.ID),
		`{"original_url": "https://strict-update.com", "short_name": "ok-link", "short_url": "boom"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "invalid request"}`, w.Body.String())
}

func TestLinkErrors(t *testing.T) {
	td := setupTestDB(t)

	const missingLinkPath = "/api/links/999"

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{name: "GetLink / invalid id", method: "GET", path: "/api/links/abc", wantStatus: http.StatusBadRequest},
		{name: "GetLink / zero id", method: http.MethodGet, path: "/api/links/0", wantStatus: http.StatusBadRequest},
		{name: "GetLink / negative id", method: http.MethodGet, path: "/api/links/-1", wantStatus: http.StatusBadRequest},
		{name: "GetLink / not found", method: "GET", path: missingLinkPath, wantStatus: http.StatusNotFound},
		{
			name:       "UpdateLink / not found",
			method:     "PUT",
			path:       missingLinkPath,
			body:       `{"original_url": "https://nowhere.com", "short_name": "ghost"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "UpdateLink / zero id",
			method:     http.MethodPut,
			path:       "/api/links/00",
			body:       `{"original_url": "https://nowhere.com", "short_name": "ghost"}`,
			wantStatus: http.StatusBadRequest,
		},
		{name: "DeleteLink / invalid id", method: "DELETE", path: "/api/links/abc", wantStatus: http.StatusBadRequest},
		{name: "DeleteLink / zero id", method: http.MethodDelete, path: "/api/links/000", wantStatus: http.StatusBadRequest},
		{name: "DeleteLink / negative id", method: http.MethodDelete, path: "/api/links/-3", wantStatus: http.StatusBadRequest},
		{name: "DeleteLink / not found", method: "DELETE", path: missingLinkPath, wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			w := performRequest(t, tx.router, tt.method, tt.path, tt.body)

			assert.Equal(t, tt.wantStatus, w.Code)
			assertErrorBody(t, w)
		})
	}
}
