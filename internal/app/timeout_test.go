package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timeoutErrorBody = `{"error": "request timeout"}`

func newTimeoutTestServer(timeout time.Duration, routes func(*gin.Engine)) *httptest.Server {
	router := gin.New()
	routes(router)

	return httptest.NewServer(http.TimeoutHandler(router, timeout, timeoutErrorBody))
}

func TestRequestTimeoutReturns503(t *testing.T) {
	srv := newTimeoutTestServer(50*time.Millisecond, func(r *gin.Engine) {
		r.GET("/slow", func(c *gin.Context) {
			time.Sleep(200 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/slow")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Contains(t, string(body), "request timeout")
}

func TestRequestTimeoutFastHandlerPassesThrough(t *testing.T) {
	srv := newTimeoutTestServer(1*time.Second, func(r *gin.Engine) {
		r.GET("/fast", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/fast")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"ok": true}`, string(body))
}

func TestRequestTimeoutForwardsDeadlineToContext(t *testing.T) {
	srv := newTimeoutTestServer(50*time.Millisecond, func(r *gin.Engine) {
		r.GET("/check", func(c *gin.Context) {
			<-c.Request.Context().Done()
			assert.Equal(t, "context deadline exceeded", c.Request.Context().Err().Error())
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/check")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
}
