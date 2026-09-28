package httpadapter

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// serveLogged runs one request through the request log and request id
// middleware, the way the app wires them, and returns the log output.
func serveLogged(t *testing.T, path string) (*httptest.ResponseRecorder, string) {
	t.Helper()

	var out bytes.Buffer
	router := gin.New()
	router.Use(RequestLog(slog.New(slog.NewTextHandler(&out, nil))))
	router.Use(RequestID())
	router.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/boom", func(c *gin.Context) {
		writeServiceError(c, errors.New("db is down"))
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

	return w, out.String()
}

func TestRequestLogInfoLine(t *testing.T) {
	w, line := serveLogged(t, "/ok")

	assert.Contains(t, line, "level=INFO")
	assert.Contains(t, line, "path=/ok")
	assert.Contains(t, line, "status=200")
	assert.Contains(t, line, "request_id="+w.Header().Get(RequestIDHeader))
}

func TestRequestLogErrorLineCarriesHiddenError(t *testing.T) {
	w, line := serveLogged(t, "/boom")

	assert.JSONEq(t, `{"error": "internal error"}`, w.Body.String())
	assert.Contains(t, line, "level=ERROR")
	assert.Contains(t, line, "status=500")
	assert.Contains(t, line, `error="db is down"`)
	assert.Contains(t, line, "request_id="+w.Header().Get(RequestIDHeader))
}
