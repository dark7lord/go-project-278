package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"code/internal/db"
)

func TestLinksCRUD(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateLink", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://example.com", "short_name": "my-link"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		tx.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("GetLink / not found", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/links/999", nil)
		tx.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("GetLink", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://gettest.com", "get-test", "http://localhost:8080/get-test")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/links/%d", created.ID), nil)
		tx.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp db.Link
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "https://gettest.com", resp.OriginalURL)
	})

	t.Run("UpdateLink", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://updatetest.com", "update-test", "http://localhost:8080/update-test")
		require.NoError(t, err)

		body := `{"original_url": "https://updated.com", "short_name": "updated-link"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/links/%d", created.ID), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		tx.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteLink", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://deletetest.com", "delete-test", "http://localhost:8080/delete-test")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/links/%d", created.ID), nil)
		tx.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("ListLinks", func(t *testing.T) {
		tx := setupTestTx(t, td)
		for i := range 10 {
			l := linkFactory(i)
			if _, err := tx.repo.CreateLink(ctx, l.OriginalURL, l.ShortName, l.ShortURL); err != nil {
				require.NoError(t, err)
			}
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/links", nil)
		tx.router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var links []db.Link
		err := json.Unmarshal(w.Body.Bytes(), &links)
		require.NoError(t, err)
		fmt.Println(len(links))
		assert.NotEqual(t, links, nil)

	})
}
