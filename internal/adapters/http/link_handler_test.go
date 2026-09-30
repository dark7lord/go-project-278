package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"code/internal/application"
)

func TestHandlerRedirectMapsVisitMetadata(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		Redirect(mock.Anything, "target", application.Visit{
			IP:        "192.0.2.1",
			UserAgent: "test-agent",
			Status:    int32(http.StatusFound),
		}).
		Return(application.LinkView{OriginalURL: testExampleURL}, nil).
		Once()
	handler := newTestHandler(linkService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/r/target", nil)
	req.Header.Set("User-Agent", "test-agent")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, testExampleURL, w.Header().Get("Location"))
}

func TestHandlerCreateLinkBuildsShortURL(t *testing.T) {
	linkService := NewMockUseCase(t)
	linkService.EXPECT().
		CreateLink(mock.Anything, application.LinkInput{
			OriginalURL: testExampleURL,
			ShortName:   "example",
		}).
		Return(application.LinkView{
			ID:          7,
			OriginalURL: testExampleURL,
			ShortName:   "example",
		}, nil).
		Once()
	handler := newTestHandler(linkService)

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
}
