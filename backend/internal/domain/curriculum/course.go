package curriculum

import (
	"strings"
	"time"

	"boilerplate-learner/internal/shared/apperrors"
)

const (
	CourseDraft     = "draft"
	CoursePublished = "published"
)

type Course struct {
	ID             int32
	OrganizationID int32
	Title          string
	Slug           string
	Description    string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Modules        []Module
}

type Module struct {
	ID       int32
	CourseID int32
	Title    string
	Position int
	Lessons  []Lesson
}

type Lesson struct {
	ID       int32
	ModuleID int32
	Title    string
	Kind     string
	Content  string
	Position int
}

func ValidateCourse(title, description string) error {
	if strings.TrimSpace(title) == "" {
		return apperrors.Invalid("el título del curso es obligatorio")
	}
	if len([]rune(title)) > 160 {
		return apperrors.Invalid("el título no puede superar 160 caracteres")
	}
	if len([]rune(description)) > 5000 {
		return apperrors.Invalid("la descripción no puede superar 5000 caracteres")
	}
	return nil
}

// Publish transitions a course to published only after it has learning content.
// Keeping the transition here prevents persistence adapters from owning policy.
func (c *Course) Publish(lessonCount int) error {
	if c.Status != CourseDraft {
		return apperrors.Conflict("el curso no está en borrador")
	}
	if lessonCount == 0 {
		return apperrors.Invalid("añade al menos una lección antes de publicar el curso")
	}
	c.Status = CoursePublished
	return nil
}

func (c Course) EnsureDraft() error {
	if c.Status != CourseDraft {
		return apperrors.Conflict("la estructura de un curso publicado no se puede modificar")
	}
	return nil
}

func ValidateModule(title string) error {
	if strings.TrimSpace(title) == "" || len([]rune(title)) > 160 {
		return apperrors.Invalid("el título del módulo es obligatorio y no puede superar 160 caracteres")
	}
	return nil
}

func ValidateLesson(title, kind string) error {
	if strings.TrimSpace(title) == "" || len([]rune(title)) > 160 {
		return apperrors.Invalid("el título de la lección es obligatorio y no puede superar 160 caracteres")
	}
	switch kind {
	case "text", "video", "link", "file", "task", "quiz":
		return nil
	default:
		return apperrors.Invalid("tipo de lección inválido")
	}
}
