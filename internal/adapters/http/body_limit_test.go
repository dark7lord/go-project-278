package httpadapter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// newBodyLimitUnitEngine builds a bare gin engine carrying only the body-limit
// middleware and a single stub API route. It needs no database: the middleware
// is the unit under test, and the stub mirrors the real links handler by
// draining the body and rejecting an over-large payload with 400.
func newBodyLimitUnitEngine(routes func(*gin.RouterGroup)) *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	api.Use(MaxRequestBody(MaxRequestBodyBytes))
	routes(api)

	return router
}

func TestRequestBodyTooLargeIsRejected(t *testing.T) {
	engine := newBodyLimitUnitEngine(func(api *gin.RouterGroup) {
		api.POST("/links", func(c *gin.Context) {
			if _, err := io.Copy(io.Discard, c.Request.Body); err != nil {
				c.Status(http.StatusBadRequest)
				return
			}
			c.Status(http.StatusNoContent)
		})
	})

	body := strings.Repeat("a", MaxRequestBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(body))

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
