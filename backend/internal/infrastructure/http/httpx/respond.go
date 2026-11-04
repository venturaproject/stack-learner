// Package httpx bridges the transport-neutral apperrors taxonomy to HTTP. It is
// the single place that knows which status code each apperrors.Kind maps to.
package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"boilerplate-learner/internal/shared/apperrors"
)

// StatusFor maps an error Kind to an HTTP status code.
func StatusFor(kind apperrors.Kind) int {
	switch kind {
	case apperrors.KindNotFound:
		return http.StatusNotFound
	case apperrors.KindInvalid:
		return http.StatusBadRequest
	case apperrors.KindConflict:
		return http.StatusConflict
	case apperrors.KindUnauthorized:
		return http.StatusUnauthorized
	case apperrors.KindForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// Status returns the HTTP status code for err based on its apperrors.Kind.
func Status(err error) int { return StatusFor(apperrors.KindOf(err)) }

// Respond writes a JSON error body with the status derived from err.
func Respond(c *gin.Context, err error) {
	c.JSON(Status(err), gin.H{"error": err.Error()})
}

// Abort is Respond + c.Abort, for use inside middleware.
func Abort(c *gin.Context, err error) {
	c.AbortWithStatusJSON(Status(err), gin.H{"error": err.Error()})
}
