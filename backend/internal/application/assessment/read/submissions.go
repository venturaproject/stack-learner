package read

import "time"

type Submission struct {
	ID             int64      `json:"id"`
	OrganizationID int32      `json:"organization_id"`
	CourseID       int32      `json:"course_id"`
	LessonID       int32      `json:"lesson_id"`
	LessonTitle    string     `json:"lesson_title"`
	StudentID      int32      `json:"student_id"`
	StudentName    string     `json:"student_name"`
	StudentEmail   string     `json:"student_email"`
	Attempt        int        `json:"attempt"`
	ResponseText   string     `json:"response_text"`
	ResponseURL    *string    `json:"response_url,omitempty"`
	Status         string     `json:"status"`
	Score          *float64   `json:"score,omitempty"`
	Feedback       *string    `json:"feedback,omitempty"`
	SubmittedAt    time.Time  `json:"submitted_at"`
	ReviewedBy     *int32     `json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
}
