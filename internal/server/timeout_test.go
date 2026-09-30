package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTimeoutTestServer(timeout time.Duration, routes func(*gin.Engine)) *httptest.Server {
	router := gin.New()
	routes(router)

	return httptest.NewServer(withRequestTimeout(router, timeout))
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
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

	var payload map[string]string
	require.NoError(t, json.Unmarshal(body, &payload))
	assert.Contains(t, payload, "error")
	assert.NotEmpty(t, payload["error"])
}

func TestRequestTimeoutKeepsDeclaredContentType(t *testing.T) {
	tests := []struct {
		name     string
		register func(*gin.Engine)
		path     string
		wantType string
	}{
		{
			name: "json response",
			register: func(r *gin.Engine) {
				r.GET("/json", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"ok": true})
				})
			},
			path:     "/json",
			wantType: "application/json; charset=utf-8",
		},
		{
			name: "redirect response",
			register: func(r *gin.Engine) {
				r.GET("/redirect", func(c *gin.Context) {
					c.Redirect(http.StatusFound, "https://example.com")
				})
			},
			path:     "/redirect",
			wantType: "text/html; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTimeoutTestServer(time.Second, tt.register)
			defer srv.Close()

			client := &http.Client{
				CheckRedirect: func(*http.Request, []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			resp, err := client.Get(srv.URL + tt.path)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.wantType, resp.Header.Get("Content-Type"))
		})
	}
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
