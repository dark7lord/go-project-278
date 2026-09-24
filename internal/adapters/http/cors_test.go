package httpadapter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newCORSUnitEngine builds a bare gin engine carrying only the CORS
// middleware and the API route whose responses must carry CORS headers.
// It needs no database and no handler logic: the middleware is the unit
// under test, so a stub route returning 200 suffices.
func newCORSUnitEngine(routes func(*gin.RouterGroup)) *gin.Engine {
	router := gin.New()
	router.Use(NewCORS())

	api := router.Group("/api")
	routes(api)

	return router
}

func TestCORSPreflightAllowsHeadersAndMethods(t *testing.T) {
	engine := newCORSUnitEngine(func(api *gin.RouterGroup) {
		api.POST("/links", func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		})
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/links", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Range")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), http.MethodDelete)
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Range")
}

func TestCORSExposesContentRange(t *testing.T) {
	engine := newCORSUnitEngine(func(api *gin.RouterGroup) {
		api.GET("/links", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Expose-Headers"), "Content-Range")
}
