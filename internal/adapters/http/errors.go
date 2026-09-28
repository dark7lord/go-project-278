package httpadapter

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"code/internal/application"
)

const errInternal = "internal error"

func errJSON(msg string) gin.H {
	return gin.H{"error": msg}
}

var camelRe = regexp.MustCompile(`([a-z0-9])([A-Z])`)

// toSnakeCase converts "ShortName" to "short_name" and "OriginalURL" to "original_url".
func toSnakeCase(s string) string {
	return strings.ToLower(camelRe.ReplaceAllString(s, `${1}_${2}`))
}

// bindMessage translates a validator tag into a human-readable message.
func bindMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	default:
		return fe.Error()
	}
}

// writeBindErrors returns 422 for validator errors, otherwise 400 for invalid JSON.
func writeBindErrors(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		result := make(map[string]string)
		for _, fieldErr := range validationErrors {
			result[toSnakeCase(fieldErr.Field())] = bindMessage(fieldErr)
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": result})

		return
	}
	c.JSON(http.StatusBadRequest, errJSON("invalid request"))
}

// writeServiceError maps an application error to a stable HTTP response.
// Only classified errors are exposed to the client; everything else is logged
// and reported as a generic 500 so internal details never leak.
func writeServiceError(c *gin.Context, err error) {
	var fe *application.FieldError
	if errors.As(err, &fe) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{fe.Field: fe.Error()}})
		return
	}

	// A classified error answers with its sentinel text: the wrap chain added
	// by the use cases is for logs, not part of the contract.
	if errors.Is(err, application.ErrSortField) {
		c.JSON(http.StatusBadRequest, errJSON(application.ErrSortField.Error()))
		return
	}

	if errors.Is(err, application.ErrNotFound) {
		c.JSON(http.StatusNotFound, errJSON(application.ErrNotFound.Error()))
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		c.JSON(http.StatusServiceUnavailable, errJSON("request timeout"))
		return
	}

	_ = c.Error(err)
	captureException(c, err)
	c.JSON(http.StatusInternalServerError, errJSON(errInternal))
}

// captureException reports err through the request's Sentry hub, so the event
// carries the request scope (request id); it falls back to the global hub.
func captureException(c *gin.Context, err error) {
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
		return
	}
	sentry.CaptureException(err)
}
