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

func TestLinksCRUD(t *testing.T) {
	td := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateLink", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://example.com", "short_name": "my-link"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("CreateLink / duplicate short name", func(t *testing.T) {
		tx := setupTestTx(t, td)
		_, err := tx.repo.CreateLink(ctx, "https://dup.com", "dup-link", "http://localhost:8080/dup-link")
		require.NoError(t, err)

		body := `{"original_url": "https://another.com", "short_name": "dup-link"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assertFieldErrors(t, w, "short_name")
	})

	t.Run("CreateLink / empty short name", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://autogen.com"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp db.Link
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ShortName)
		assert.Contains(t, resp.ShortURL, "http://localhost:8080/")
	})

	t.Run("CreateLink / missing original_url", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "", "short_name": "ok-link"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assertFieldErrors(t, w, "original_url")
	})

	t.Run("CreateLink / short name too short", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://short.com", "short_name": "x"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assertFieldErrors(t, w, "short_name")
	})

	t.Run("CreateLink / short name too long", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://long.com", "short_name": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
		assertFieldErrors(t, w, "short_name")
	})

	t.Run("CreateLink / scheme-less url", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "example.com", "short_name": "no-scheme"}`
		w := performRequest(t, tx.router, "POST", "/api/links", body)
		assert.Equal(t, http.StatusCreated, w.Code)

		var resp db.Link
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", resp.OriginalURL)
		assert.Contains(t, resp.ShortURL, "/r/no-scheme")
	})

	t.Run("CreateLink / invalid json", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := performRequest(t, tx.router, "POST", "/api/links", `{broken`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var body map[string]string
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, "invalid request", body["error"])
	})

	t.Run("GetLink / not found", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := performRequest(t, tx.router, "GET", "/api/links/999", "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assertErrorBody(t, w)
	})

	t.Run("GetLink / invalid id", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := performRequest(t, tx.router, "GET", "/api/links/abc", "")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertErrorBody(t, w)
	})

	t.Run("GetLink", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://gettest.com", "get-test", "http://localhost:8080/get-test")
		require.NoError(t, err)

		w := performRequest(t, tx.router, "GET", fmt.Sprintf("/api/links/%d", created.ID), "")
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
		w := performRequest(t, tx.router, "PUT", fmt.Sprintf("/api/links/%d", created.ID), body)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateLink / not found", func(t *testing.T) {
		tx := setupTestTx(t, td)
		body := `{"original_url": "https://nowhere.com", "short_name": "ghost"}`
		w := performRequest(t, tx.router, "PUT", "/api/links/999", body)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assertErrorBody(t, w)
	})

	t.Run("UpdateLink / duplicate short name", func(t *testing.T) {
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

	t.Run("UpdateLink / same short_name", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, _ := tx.repo.CreateLink(ctx, "https://a.com", "same", "http://localhost:8080/r/same")
		body := `{"original_url": "https://b.com", "short_name": "same"}`
		w := performRequest(t, tx.router, "PUT", fmt.Sprintf("/api/links/%d", created.ID), body)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("DeleteLink", func(t *testing.T) {
		tx := setupTestTx(t, td)
		created, err := tx.repo.CreateLink(ctx, "https://deletetest.com", "delete-test", "http://localhost:8080/delete-test")
		require.NoError(t, err)

		w := performRequest(t, tx.router, "DELETE", fmt.Sprintf("/api/links/%d", created.ID), "")
		assert.Equal(t, http.StatusOK, w.Code)

		var resp db.Link
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, created.ID, resp.ID)
		assert.Equal(t, "https://deletetest.com", resp.OriginalURL)
	})

	t.Run("DeleteLink / invalid id", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := performRequest(t, tx.router, "DELETE", "/api/links/abc", "")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assertErrorBody(t, w)
	})

	t.Run("DeleteLink / not found", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := performRequest(t, tx.router, "DELETE", "/api/links/999", "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assertErrorBody(t, w)
	})

	t.Run("ListLinks", func(t *testing.T) {
		tx := setupTestTx(t, td)
		for i := range 10 {
			l := linkFactory(i)
			_, err := tx.repo.CreateLink(ctx, l.OriginalURL, l.ShortName, l.ShortURL)
			require.NoError(t, err)
		}

		w := performRequest(t, tx.router, "GET", "/api/links", "")
		assert.Equal(t, http.StatusOK, w.Code)

		var links []db.Link
		err := json.Unmarshal(w.Body.Bytes(), &links)
		require.NoError(t, err)
		assert.Len(t, links, 10)
	})

	t.Run("ListLinks / empty returns []", func(t *testing.T) {
		tx := setupTestTx(t, td)
		w := performRequest(t, tx.router, "GET", "/api/links", "")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "[]", w.Body.String())
	})
}
