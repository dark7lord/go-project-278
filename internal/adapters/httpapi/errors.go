package httpapi

import (
	"context"
	"errors"
	"net/http"

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
	if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
		result := make(map[string]string)
		for _, fieldErr := range validationErrors {
			result[fieldErr.Field()] = bindMessage(fieldErr)
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": result})

		return
	}
	c.JSON(http.StatusBadRequest, errJSON("invalid request"))
}

// writeServiceError answers a known application error with its own text and
// hides anything else behind a logged 500: wrap chains are for logs only.
func writeServiceError(c *gin.Context, err error) {
	if fe, ok := errors.AsType[*application.FieldError](err); ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{fe.Field: fe.Error()}})
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

	reportError(c, err)
	c.JSON(http.StatusInternalServerError, errJSON(errInternal))
}

// reportError logs err and sends it to Sentry without answering the client.
func reportError(c *gin.Context, err error) {
	_ = c.Error(err)
	captureException(c, err)
}

// captureException reports err to the request's Sentry hub, so the event
// carries the request id, or to the global hub without one.
func captureException(c *gin.Context, err error) {
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.CaptureException(err)
		return
	}
	sentry.CaptureException(err)
}
