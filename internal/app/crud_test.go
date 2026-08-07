package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/db"
)

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
			wantField:  "original_url",
		},
		{
			name:       "short name too short",
			body:       `{"original_url": "https://short.com", "short_name": "x"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  "short_name",
		},
		{
			name:       "short name too long",
			body:       `{"original_url": "https://long.com", "short_name": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`,
			wantStatus: http.StatusUnprocessableEntity,
			wantField:  "short_name",
		},
		{
			name:       "invalid json",
			body:       `{broken`,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := setupTestTx(t, td)

			w := performRequest(t, tx.router, "POST", "/api/links", tt.body)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantField != "" {
				assertFieldErrors(t, w, tt.wantField)
			} else {
				assert.JSONEq(t, `{"error": "`+tt.wantError+`"}`, w.Body.String())
			}
		})
	}
}

func TestCreateLink(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://example.com", "short_name": "my-link"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("empty short name", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://autogen.com"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp db.Link
		decode(t, w, &resp)
		assert.NotEmpty(t, resp.ShortName)
		assert.Contains(t, resp.ShortURL, "http://localhost:8080/")
	})

	t.Run("scheme-less url", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "example.com", "short_name": "no-scheme"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp db.Link
		decode(t, w, &resp)
		assert.Equal(t, "https://example.com", resp.OriginalURL)
		assert.Contains(t, resp.ShortURL, "/r/no-scheme")
	})

	t.Run("duplicate short name", func(t *testing.T) {
		tx := setupTestTx(t, td)
		_, err := tx.repo.CreateLink(ctx, "https://dup.com", "dup-link", "http://localhost:8080/dup-link")
		require.NoError(t, err)

		body := `{"original_url": "https://another.com", "short_name": "dup-link"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assertFieldErrors(t, w, "short_name")
	})
}

func TestGetLink(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://gettest.com", "get-test", "http://localhost:8080/get-test")
		require.NoError(t, err)

		w := performRequest(t, tx.router, "GET", fmt.Sprintf("/api/links/%d", created.ID), "")
		assert.Equal(t, http.StatusOK, w.Code)

		expected, err := json.Marshal(created)
		require.NoError(t, err)
		assert.JSONEq(t, string(expected), w.Body.String())
	})
}

func TestUpdateLink(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://updatetest.com", "update-test", "http://localhost:8080/update-test")
		require.NoError(t, err)

		body := `{"original_url": "https://updated.com", "short_name": "updated-link"}`
		w := performRequest(t, tx.router, "PUT", fmt.Sprintf("/api/links/%d", created.ID), body)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("duplicate short name", func(t *testing.T) {
		tx := setupTestTx(t, td)
		_, err := tx.repo.CreateLink(ctx, "https://taken.com", "taken", "http://localhost:8080/taken")
		require.NoError(t, err)
		other, err := tx.repo.CreateLink(ctx, "https://other.com", "other", "http://localhost:8080/other")
		require.NoError(t, err)

		body := `{"original_url": "https://other.com", "short_name": "taken"}`
		w := performRequest(t, tx.router, "PUT", fmt.Sprintf("/api/links/%d", other.ID), body)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assertFieldErrors(t, w, "short_name")
	})

	t.Run("same short_name", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, _ := tx.repo.CreateLink(ctx, "https://a.com", "same", "http://localhost:8080/r/same")
		body := `{"original_url": "https://b.com", "short_name": "same"}`
		w := performRequest(t, tx.router, "PUT", fmt.Sprintf("/api/links/%d", created.ID), body)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestDeleteLink(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("ok", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://deletetest.com", "delete-test", "http://localhost:8080/delete-test")
		require.NoError(t, err)

		w := performRequest(t, tx.router, "DELETE", fmt.Sprintf("/api/links/%d", created.ID), "")
		assert.Equal(t, http.StatusOK, w.Code)

		expected, err := json.Marshal(created)
		require.NoError(t, err)
		assert.JSONEq(t, string(expected), w.Body.String())
	})
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
		{name: "GetLink / not found", method: "GET", path: missingLinkPath, wantStatus: http.StatusNotFound},
		{
			name:       "UpdateLink / not found",
			method:     "PUT",
			path:       missingLinkPath,
			body:       `{"original_url": "https://nowhere.com", "short_name": "ghost"}`,
			wantStatus: http.StatusNotFound,
		},
		{name: "DeleteLink / invalid id", method: "DELETE", path: "/api/links/abc", wantStatus: http.StatusBadRequest},
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

func TestListLinks(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("returns all", func(t *testing.T) {
		tx := setupTestTx(t, td)
		for i := range 10 {
			l := linkFactory(i)
			_, err := tx.repo.CreateLink(ctx, l.OriginalURL, l.ShortName, l.ShortURL)
			require.NoError(t, err)
		}

		w := performRequest(t, tx.router, "GET", "/api/links", "")
		assert.Equal(t, http.StatusOK, w.Code)

		var links []db.Link
		decode(t, w, &links)
		assert.Len(t, links, 10)
	})

	t.Run("empty returns []", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := performRequest(t, tx.router, "GET", "/api/links", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", w.Body.String())
	})
}
