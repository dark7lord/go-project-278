package httpapi

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newCORSUnitEngine builds a bare engine with the CORS middleware in front of
// a stub API route.
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

// newRequestIDUnitEngine puts the request id middleware after sentrygin, as the
// app does, in front of stub routes.
func newRequestIDUnitEngine() *gin.Engine {
	router := gin.New()
	router.Use(sentrygin.New(sentrygin.Options{}))
	router.Use(RequestID())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/boom", func(c *gin.Context) {
		writeServiceError(c, errors.New("boom"))
	})

	return router
}

func TestRequestIDInResponse(t *testing.T) {
	w := httptest.NewRecorder()
	newRequestIDUnitEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))

	assert.NotEmpty(t, w.Header().Get(requestIDHeader))
}

func TestRequestIDTagsSentryEvent(t *testing.T) {
	var events []*sentry.Event
	err := sentry.Init(sentry.ClientOptions{
		BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			events = append(events, event)
			return nil
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { sentry.CurrentHub().BindClient(nil) })

	w := httptest.NewRecorder()
	newRequestIDUnitEngine().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/boom", nil))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.Len(t, events, 1)
	assert.Equal(t, w.Header().Get(requestIDHeader), events[0].Tags["request_id"])
}

// serveLogged runs one request through the request log and request id
// middleware, wired as in the app, and returns the log line.
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
	router.GET("/quiet", func(c *gin.Context) {
		reportError(c, errors.New("visit not recorded"))
		c.Status(http.StatusFound)
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
	assert.Contains(t, line, "request_id="+w.Header().Get(requestIDHeader))
}

func TestRequestLogErrorLineCarriesHiddenError(t *testing.T) {
	w, line := serveLogged(t, "/boom")

	assert.JSONEq(t, `{"error": "internal error"}`, w.Body.String())
	assert.Contains(t, line, "level=ERROR")
	assert.Contains(t, line, "status=500")
	assert.Contains(t, line, `error="db is down"`)
	assert.Contains(t, line, "request_id="+w.Header().Get(requestIDHeader))
}

func TestRequestLogWarnLineCarriesReportedError(t *testing.T) {
	w, line := serveLogged(t, "/quiet")

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, line, "level=WARN")
	assert.Contains(t, line, "status=302")
	assert.Contains(t, line, `error="visit not recorded"`)
}
