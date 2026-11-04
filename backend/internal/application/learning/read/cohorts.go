package read

import "time"

type Cohort struct {
	ID              int32      `json:"id"`
	OrganizationID  int32      `json:"organization_id"`
	CourseID        int32      `json:"course_id"`
	CourseTitle     string     `json:"course_title"`
	Name            string     `json:"name"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	Status          string     `json:"status"`
	EnrollmentCount int64      `json:"enrollment_count"`
}

type CohortInstructor struct {
	UserID   int32  `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Assigned bool   `json:"assigned"`
}

type CohortEnrollment struct {
	ID               int64      `json:"id"`
	CohortID         int32      `json:"cohort_id"`
	UserID           int32      `json:"user_id"`
	Name             string     `json:"name"`
	Email            string     `json:"email"`
	Status           string     `json:"status"`
	EnrolledAt       time.Time  `json:"enrolled_at"`
	WithdrawnAt      *time.Time `json:"withdrawn_at,omitempty"`
	WithdrawalReason *string    `json:"withdrawal_reason,omitempty"`
}

type LessonProgress struct {
	LessonID    int32      `json:"lesson_id"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Unlocked    bool       `json:"unlocked"`
}

type CourseProgress struct {
	CourseID         int32            `json:"course_id"`
	TotalLessons     int64            `json:"total_lessons"`
	CompletedLessons int64            `json:"completed_lessons"`
	Percent          int              `json:"percent"`
	Lessons          []LessonProgress `json:"lessons"`
}
