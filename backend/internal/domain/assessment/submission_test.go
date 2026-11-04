package assessment

import (
	"testing"
	"time"

	"boilerplate-learner/internal/shared/apperrors"
)

func TestValidateSubmissionResponse(t *testing.T) {
	if err := ValidateResponse("", ""); apperrors.KindOf(err) != apperrors.KindInvalid {
		t.Fatalf("empty submission must be rejected, got %v", err)
	}
	if err := ValidateResponse("answer", "javascript:alert(1)"); apperrors.KindOf(err) != apperrors.KindInvalid {
		t.Fatalf("unsafe URL must be rejected, got %v", err)
	}
	if err := ValidateResponse("answer", "https://example.test/work"); err != nil {
		t.Fatalf("valid text and URL should pass validation: %v", err)
	}
}

func TestReviewSubmissionValidatesAndTransitionsOnce(t *testing.T) {
	submission := Submission{Status: StatusSubmitted}
	if err := submission.Review(9, 101, "", time.Now()); apperrors.KindOf(err) != apperrors.KindInvalid {
		t.Fatalf("score above 100 must fail, got %v", err)
	}
	if err := submission.Review(9, 88.5, "Good work", time.Now()); err != nil {
		t.Fatal(err)
	}
	if submission.Status != StatusReviewed || submission.Score == nil || *submission.Score != 88.5 || submission.Feedback == nil {
		t.Fatalf("review state not stored: %+v", submission)
	}
	if err := submission.Review(9, 90, "Second review", time.Now()); apperrors.KindOf(err) != apperrors.KindConflict {
		t.Fatalf("reviewed submission must not be overwritten, got %v", err)
	}
}
