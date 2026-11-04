package cohorts

import (
	"strings"
	"time"

	"boilerplate-learner/internal/shared/apperrors"
)

const (
	StatusDraft         = "draft"
	StatusActive        = "active"
	StatusCompleted     = "completed"
	StatusCancelled     = "cancelled"
	EnrollmentActive    = "active"
	EnrollmentWithdrawn = "withdrawn"
)

type Cohort struct {
	ID             int32
	OrganizationID int32
	CourseID       int32
	Name           string
	StartsAt       *time.Time
	EndsAt         *time.Time
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func Validate(name string, startsAt, endsAt *time.Time) error {
	if strings.TrimSpace(name) == "" || len([]rune(name)) > 160 {
		return apperrors.Invalid("el nombre de la cohorte es obligatorio y no puede superar 160 caracteres")
	}
	if startsAt != nil && endsAt != nil && endsAt.Before(*startsAt) {
		return apperrors.Invalid("la fecha de finalización no puede ser anterior a la fecha de inicio")
	}
	return nil
}

func CanTransition(from, to string) bool {
	switch from {
	case StatusDraft:
		return to == StatusActive || to == StatusCancelled
	case StatusActive:
		return to == StatusCompleted || to == StatusCancelled
	default:
		return false
	}
}
