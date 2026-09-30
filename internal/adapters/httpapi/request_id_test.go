package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRequestIDUnitEngine wires the request id middleware the way the app does,
// after sentrygin, in front of stub routes.
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

	assert.NotEmpty(t, w.Header().Get(RequestIDHeader))
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
	assert.Equal(t, w.Header().Get(RequestIDHeader), events[0].Tags["request_id"])
}
