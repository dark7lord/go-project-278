package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"code/internal/config"
)

func TestNewServerKeepsWriteTimeoutAboveRequestTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{name: "default", timeout: 10 * time.Second},
		{name: "fast", timeout: 250 * time.Millisecond},
		{name: "slow", timeout: 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{RequestTimeout: tt.timeout}

			server := newServer(cfg, http.NotFoundHandler())

			assert.Greater(t, server.WriteTimeout, cfg.RequestTimeout)
			assert.Equal(t, cfg.RequestTimeout+writeTimeoutGrace, server.WriteTimeout)
		})
	}
}

func TestNewServerWiresRequestTimeout(t *testing.T) {
	cfg := &config.Config{RequestTimeout: 50 * time.Millisecond}
	router := gin.New()
	router.GET("/slow", func(c *gin.Context) {
		<-c.Request.Context().Done()
	})

	server := newServer(cfg, router)

	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/slow", nil))

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestNewServerLimitsRequestBody(t *testing.T) {
	router := gin.New()
	router.POST("/api/links", func(c *gin.Context) {
		if _, err := io.Copy(io.Discard, c.Request.Body); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		c.Status(http.StatusNoContent)
	})

	server := newServer(&config.Config{RequestTimeout: time.Second}, router)

	body := strings.NewReader(strings.Repeat("a", maxRequestBodyBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/api/links", body)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
