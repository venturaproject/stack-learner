package httpx

import (
	"net/http"
	"testing"

	"boilerplate-learner/internal/shared/apperrors"
)

func TestStatusForMapsEveryKind(t *testing.T) {
	cases := map[error]int{
		apperrors.NotFound("x"):     http.StatusNotFound,
		apperrors.Invalid("x"):      http.StatusBadRequest,
		apperrors.Conflict("x"):     http.StatusConflict,
		apperrors.Unauthorized("x"): http.StatusUnauthorized,
		apperrors.Forbidden("x"):    http.StatusForbidden,
		apperrors.Internal("x"):     http.StatusInternalServerError,
	}
	for err, want := range cases {
		if got := Status(err); got != want {
			t.Errorf("Status(%v) = %d, want %d", err, got, want)
		}
	}
}

func TestStatusForUnknownErrorIsInternal(t *testing.T) {
	if got := Status(http.ErrAbortHandler); got != http.StatusInternalServerError {
		t.Errorf("unknown error mapped to %d, want 500", got)
	}
}

func TestStatusForWrappedError(t *testing.T) {
	wrapped := apperrors.Wrap(apperrors.KindConflict, "duplicate", http.ErrBodyNotAllowed)
	if got := Status(wrapped); got != http.StatusConflict {
		t.Errorf("wrapped conflict mapped to %d, want 409", got)
	}
}
