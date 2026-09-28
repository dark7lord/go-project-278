package httpadapter

import (
	"crypto/rand"
	"log/slog"
	"net/http"
	"strings"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// The decoder flag is process-global, so the transport that decodes bodies owns it.
func init() {
	binding.EnableDecoderDisallowUnknownFields = true
}

// MaxRequestBodyBytes bounds JSON request bodies from above (1 MiB headroom).
const MaxRequestBodyBytes = 1 << 20

// MaxRequestBody is the equivalent of the common gin.MaxAllowedBodyBytes
// helper (gin does not ship one): it caps the body size seen by handlers.
func MaxRequestBody(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	}
}

// NewCORS returns the CORS middleware used by the API transport.
func NewCORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodHead,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders:  []string{"Content-Type", "Accept", "Range"},
		ExposeHeaders: []string{"Content-Range"},
		MaxAge:        12 * time.Hour,
	})
}

// RequestIDHeader carries the id that ties a response to its Sentry event.
const RequestIDHeader = "X-Request-ID"

// RequestID tags each request with a fresh id in the response and its Sentry
// scope. It must run after the sentrygin middleware.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := rand.Text()
		c.Header(RequestIDHeader, id)
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			hub.Scope().SetTag("request_id", id)
		}
	}
}

// RequestLog writes one line per request with its request id. A 5xx is logged
// as an error together with the internal errors the client never sees. It
// must run first, so the latency covers the whole chain.
func RequestLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency", time.Since(start),
			"ip", c.ClientIP(),
			"request_id", c.Writer.Header().Get(RequestIDHeader),
		}

		if status < http.StatusInternalServerError {
			logger.Info("request", attrs...)
			return
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "error", strings.Join(c.Errors.Errors(), "; "))
		}
		logger.Error("request", attrs...)
	}
}
