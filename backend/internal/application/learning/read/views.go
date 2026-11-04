// Package read contains the HTTP-facing projections for the learning use cases.
package read

import (
	"time"

	"boilerplate-learner/internal/domain/curriculum"
	"boilerplate-learner/internal/domain/organizations"
)

type Organization struct {
	ID           int32      `json:"id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	Role         string     `json:"role"`
	ArchivedAt   *time.Time `json:"archived_at,omitempty"`
	CourseCount  int64      `json:"course_count"`
	StudentCount int64      `json:"student_count"`
}
type Membership struct {
	OrganizationID int32  `json:"organization_id"`
	UserID         int32  `json:"user_id"`
	Role           string `json:"role"`
}

type Student struct {
	UserID int32  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
type CourseTrainer struct {
	UserID   int32  `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Assigned bool   `json:"assigned"`
}
type Lesson struct {
	ID       int32  `json:"id"`
	ModuleID int32  `json:"module_id"`
	Title    string `json:"title"`
	Kind     string `json:"kind"`
	Content  string `json:"content"`
	Position int    `json:"position"`
}
type Module struct {
	ID       int32    `json:"id"`
	CourseID int32    `json:"course_id"`
	Title    string   `json:"title"`
	Position int      `json:"position"`
	Lessons  []Lesson `json:"lessons,omitempty"`
}
type Course struct {
	ID             int32     `json:"id"`
	OrganizationID int32     `json:"organization_id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Modules        []Module  `json:"modules,omitempty"`
}

func OrganizationFromDomain(v *organizations.Organization) *Organization {
	if v == nil {
		return nil
	}
	return &Organization{ID: v.ID, Name: v.Name, Slug: v.Slug}
}
func MembershipFromDomain(v *organizations.Membership) *Membership {
	if v == nil {
		return nil
	}
	return &Membership{OrganizationID: v.OrganizationID, UserID: v.UserID, Role: v.Role}
}
func CourseFromDomain(v *curriculum.Course) *Course {
	if v == nil {
		return nil
	}
	out := &Course{ID: v.ID, OrganizationID: v.OrganizationID, Title: v.Title, Slug: v.Slug, Description: v.Description, Status: v.Status, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
	for _, m := range v.Modules {
		module := Module{ID: m.ID, CourseID: m.CourseID, Title: m.Title, Position: m.Position}
		for _, l := range m.Lessons {
			module.Lessons = append(module.Lessons, Lesson{ID: l.ID, ModuleID: l.ModuleID, Title: l.Title, Kind: l.Kind, Content: l.Content, Position: l.Position})
		}
		out.Modules = append(out.Modules, module)
	}
	return out
}
func CoursesFromDomain(values []curriculum.Course) []Course {
	out := make([]Course, 0, len(values))
	for i := range values {
		out = append(out, *CourseFromDomain(&values[i]))
	}
	return out
}
func ModuleFromDomain(v *curriculum.Module) *Module {
	if v == nil {
		return nil
	}
	return &Module{ID: v.ID, CourseID: v.CourseID, Title: v.Title, Position: v.Position}
}
func LessonFromDomain(v *curriculum.Lesson) *Lesson {
	if v == nil {
		return nil
	}
	return &Lesson{ID: v.ID, ModuleID: v.ModuleID, Title: v.Title, Kind: v.Kind, Content: v.Content, Position: v.Position}
}
