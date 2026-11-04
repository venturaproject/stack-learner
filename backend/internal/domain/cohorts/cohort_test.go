package cohorts

import (
	"testing"
	"time"
)

func TestValidateCohortDates(t *testing.T) {
	start := time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC)
	end := start.Add(-time.Hour)
	if err := Validate("Cohort", &start, &end); err == nil {
		t.Fatal("expected end-before-start to be rejected")
	}
}

func TestCohortTransitions(t *testing.T) {
	allowed := [][2]string{{StatusDraft, StatusActive}, {StatusDraft, StatusCancelled}, {StatusActive, StatusCompleted}, {StatusActive, StatusCancelled}}
	for _, transition := range allowed {
		if !CanTransition(transition[0], transition[1]) {
			t.Errorf("transition %s -> %s should be allowed", transition[0], transition[1])
		}
	}
	if CanTransition(StatusCompleted, StatusActive) || CanTransition(StatusCancelled, StatusActive) {
		t.Fatal("terminal cohort status must not be reactivated")
	}
}
