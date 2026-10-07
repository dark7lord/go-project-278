package httpapi

import (
	"crypto/rand"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Unknown JSON fields are rejected and validation errors name fields by json
// tag; both settings are process-global, so the package that decodes owns them.
func init() {
	binding.EnableDecoderDisallowUnknownFields = true
	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validate.RegisterTagNameFunc(jsonFieldName)
	}
}

// jsonFieldName names a struct field by its json tag; an untagged field keeps its Go name.
func jsonFieldName(field reflect.StructField) string {
	name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if name == "-" {
		return ""
	}

	return name
}

// NewCORS lets the Vite dev server on :5173 call the API: it proxies /api but
// forwards its own Origin, which a write request carries and Host no longer matches.
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

// requestIDHeader carries the id that ties a response to its Sentry event.
const requestIDHeader = "X-Request-ID"

// RequestID tags each request with a fresh id in the response and its Sentry
// scope. It must run after the sentrygin middleware.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := rand.Text()
		c.Header(requestIDHeader, id)
		if hub := sentrygin.GetHubFromContext(c); hub != nil {
			hub.Scope().SetTag("request_id", id)
		}
	}
}

// RequestLog writes one line per request with its id and hidden errors: error
// level for a 5xx, warn for other errors. It must run first to time the chain.
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
			"request_id", c.Writer.Header().Get(requestIDHeader),
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "error", strings.Join(c.Errors.Errors(), "; "))
		}

		switch {
		case status >= http.StatusInternalServerError:
			logger.Error("request", attrs...)
		case len(c.Errors) > 0:
			logger.Warn("request", attrs...)
		default:
			logger.Info("request", attrs...)
		}
	}
}
