package httpadapter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"code/internal/application"
)

func newHandlerRouter(handler *Handler) *gin.Engine {
	router := gin.New()
	router.POST("/links", handler.CreateLink)
	router.GET("/links/:id", handler.GetLink)
	router.GET("/links", handler.ListLinks)
	router.GET("/r/:code", handler.Redirect)
	router.GET("/visits", handler.ListVisits)

	return router
}

func TestHandlerCreateLinkValidationDoesNotCallUseCase(t *testing.T) {
	linkService := &mockLinkUseCase{}
	handler := NewHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/links", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	linkService.AssertNotCalled(t, "CreateLink", mock.Anything, mock.Anything)
}

func TestHandlerListLinksRangeMapsRequest(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{Start: 5, End: 9}).
		Return(application.RangePage[application.LinkView]{Items: make([]application.LinkView, 5), Start: 5, Total: 10}, nil).
		Once()
	handler := NewHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?range=%5B5%2C9%5D", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "links 5-9/10", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksRangeUnsatisfiable(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{Start: 100, End: 200}).
		Return(application.RangePage[application.LinkView]{Items: []application.LinkView{}, Start: 100, Total: 5}, nil).
		Once()
	handler := NewHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?range=%5B100%2C200%5D", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "links */5", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
}

func TestHandlerListLinksRangeEmptyCollection(t *testing.T) {
	linkService := &mockLinkUseCase{}
	linkService.
		On("ListLinksRange", mock.Anything, application.ListLinksQuery{Start: 0, End: 4}).
		Return(application.RangePage[application.LinkView]{Items: []application.LinkView{}, Start: 0, Total: 0}, nil).
		Once()
	handler := NewHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?range=%5B0%2C4%5D", nil)
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
	handler := NewHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links?range=%5B5%2C9%5D", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestedRangeNotSatisfiable, w.Code)
	assert.Equal(t, "links */10", w.Header().Get("Content-Range"))
	linkService.AssertExpectations(t)
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
		Return(application.LinkView{OriginalURL: "https://example.com"}, nil).
		Once()
	handler := NewHandler(linkService, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/r/target", nil)
	req.Header.Set("User-Agent", "test-agent")
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
	linkService.AssertExpectations(t)
}

func TestHandlerListVisitsRangeMapsRequest(t *testing.T) {
	visitService := &mockVisitUseCase{}
	visitService.
		On("ListLinkVisitsRange", mock.Anything, application.ListLinkVisitsQuery{Start: 0, End: 4}).
		Return(application.RangePage[application.VisitView]{Items: make([]application.VisitView, 5), Start: 0, Total: 5}, nil).
		Once()
	handler := NewHandler(&mockLinkUseCase{}, visitService)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/visits?range=%5B0%2C4%5D", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "link_visits 0-4/5", w.Header().Get("Content-Range"))
	visitService.AssertExpectations(t)
}

func TestHandlerGetLinkRejectsInvalidID(t *testing.T) {
	handler := NewHandler(&mockLinkUseCase{}, &mockVisitUseCase{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/links/not-an-id", nil)
	newHandlerRouter(handler).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, errInvalidID, body["error"])
}
