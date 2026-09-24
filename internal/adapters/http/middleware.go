package httpadapter

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

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
