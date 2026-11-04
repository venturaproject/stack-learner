package organizations

import (
	"strings"
	"unicode"

	"boilerplate-learner/internal/shared/apperrors"
)

type Organization struct {
	ID   int32
	Name string
	Slug string
}

type Membership struct {
	OrganizationID int32
	UserID         int32
	Role           string
}

const (
	RoleAdmin   = "org_admin"
	RoleTrainer = "trainer"
	RoleStudent = "student"
)

func NormalizeSlug(value string) string {
	var b strings.Builder
	separator := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if separator && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			separator = false
		} else {
			separator = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func ValidateRole(role string) error {
	switch role {
	case RoleAdmin, RoleTrainer, RoleStudent:
		return nil
	default:
		return apperrors.Invalid("rol inválido; usa org_admin, trainer o student")
	}
}
