package persistence

import (
	"context"
	"errors"
	"strings"
	"time"

	appassessment "boilerplate-learner/internal/application/assessment"
	domainassessment "boilerplate-learner/internal/domain/assessment"
	"boilerplate-learner/internal/domain/curriculum"
	"boilerplate-learner/internal/infrastructure/persistence/models"
	"boilerplate-learner/internal/shared/apperrors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AssessmentRepository struct{ db *gorm.DB }

func NewAssessmentRepository(db *gorm.DB) *AssessmentRepository {
	return &AssessmentRepository{db: db}
}

func (r *AssessmentRepository) ActiveOrganizationRole(ctx context.Context, userID, orgID int32) (string, error) {
	var role string
	err := connFromContext(ctx, r.db).Table("organization_memberships m").
		Select("m.role").Joins("JOIN organizations o ON o.id = m.organization_id").
		Where("m.user_id = ? AND m.organization_id = ? AND m.status = 'active' AND o.archived_at IS NULL", userID, orgID).
		Take(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", apperrors.NotFound("organización o membresía no encontrada")
	}
	if err != nil {
		return "", apperrors.Internal("no se pudo validar la membresía")
	}
	return role, nil
}

func (r *AssessmentRepository) CourseTrainerAssigned(ctx context.Context, orgID, courseID, trainerID int32) (bool, error) {
	var count int64
	err := connFromContext(ctx, r.db).Table("course_trainers ct").
		Joins("JOIN courses c ON c.id = ct.course_id").
		Where("c.organization_id = ? AND ct.course_id = ? AND ct.trainer_id = ?", orgID, courseID, trainerID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.Internal("no se pudo validar la asignación del docente")
	}
	return count > 0, nil
}

func (r *AssessmentRepository) ValidateStudentSubmissionAccess(ctx context.Context, orgID, courseID, lessonID, studentID int32) error {
	db := connFromContext(ctx, r.db)
	var count int64
	err := db.Table("lessons l").
		Joins("JOIN course_modules cm ON cm.id = l.module_id").
		Joins("JOIN courses c ON c.id = cm.course_id").
		Where(`l.id = ? AND l.kind = 'task' AND c.id = ? AND c.organization_id = ? AND c.status = ?
		AND EXISTS (
			SELECT 1 FROM cohorts cohort
			JOIN cohort_enrollments ce ON ce.cohort_id = cohort.id
			WHERE cohort.course_id = c.id AND cohort.organization_id = c.organization_id
			AND cohort.status IN ('active', 'completed') AND ce.user_id = ? AND ce.status = 'active'
		)`, lessonID, courseID, orgID, curriculum.CoursePublished, studentID).
		Count(&count).Error
	if err != nil {
		return apperrors.Internal("no se pudo validar el acceso a la actividad")
	}
	if count == 0 {
		return apperrors.NotFound("actividad no encontrada")
	}
	var pendingPrevious int64
	err = db.Table("lessons previous_lesson").
		Joins("JOIN course_modules previous_module ON previous_module.id = previous_lesson.module_id").
		Joins("JOIN course_modules target_module ON target_module.course_id = previous_module.course_id").
		Joins("JOIN lessons target_lesson ON target_lesson.module_id = target_module.id").
		Where(`target_lesson.id = ? AND previous_module.course_id = ? AND (
			previous_module.position < target_module.position OR
			(previous_module.position = target_module.position AND previous_lesson.position < target_lesson.position)
		) AND NOT EXISTS (
			SELECT 1 FROM student_lesson_progress progress
			WHERE progress.lesson_id = previous_lesson.id AND progress.user_id = ?
		)`, lessonID, courseID, studentID).
		Count(&pendingPrevious).Error
	if err != nil {
		return apperrors.Internal("no se pudo validar el desbloqueo de la actividad")
	}
	if pendingPrevious > 0 {
		return apperrors.Conflict("completa las lecciones anteriores para acceder a esta actividad")
	}
	return nil
}

func (r *AssessmentRepository) CreateSubmission(ctx context.Context, orgID, courseID, lessonID, studentID int32, responseText string, responseURL *string) (*domainassessment.Submission, error) {
	db := connFromContext(ctx, r.db)
	var lockedLesson struct{ ID int32 }
	if err := db.Table("lessons").Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id = ?", lessonID).Take(&lockedLesson).Error; err != nil {
		return nil, apperrors.NotFound("actividad no encontrada")
	}
	var lastAttempt int
	if err := db.Model(&models.AssessmentSubmission{}).Where("student_id = ? AND lesson_id = ?", studentID, lessonID).Select("COALESCE(MAX(attempt_number), 0)").Scan(&lastAttempt).Error; err != nil {
		return nil, apperrors.Internal("no se pudo calcular el número de intento")
	}
	row := models.AssessmentSubmission{
		OrganizationID: orgID, CourseID: courseID, LessonID: lessonID,
		StudentID: studentID, Attempt: lastAttempt + 1, ResponseText: responseText,
		ResponseURL: responseURL, Status: domainassessment.StatusSubmitted,
	}
	if err := db.Create(&row).Error; err != nil {
		return nil, apperrors.Internal("no se pudo guardar la entrega")
	}
	return r.submissionByID(db, orgID, courseID, row.ID)
}

func (r *AssessmentRepository) ListStudentSubmissions(ctx context.Context, orgID, courseID, lessonID, studentID int32) ([]domainassessment.Submission, error) {
	db := connFromContext(ctx, r.db)
	rows := make([]assessmentSubmissionRow, 0)
	err := submissionSelect(db).
		Where("s.organization_id = ? AND s.course_id = ? AND s.lesson_id = ? AND s.student_id = ?", orgID, courseID, lessonID, studentID).
		Order("s.attempt_number DESC").Scan(&rows).Error
	if err != nil {
		return nil, apperrors.Internal("no se pudo cargar el historial de entregas")
	}
	return mapAssessmentSubmissions(rows), nil
}

func (r *AssessmentRepository) ListCourseSubmissions(ctx context.Context, orgID, courseID, lessonID int32, page appassessment.PageQuery) ([]domainassessment.Submission, int64, error) {
	db := connFromContext(ctx, r.db)
	query := submissionSelect(db).Where("s.organization_id = ? AND s.course_id = ?", orgID, courseID)
	if lessonID > 0 {
		query = query.Where("s.lesson_id = ?", lessonID)
	}
	if page.Search != "" {
		like := "%" + strings.TrimSpace(page.Search) + "%"
		query = query.Where("(u.name ILIKE ? OR u.email ILIKE ? OR l.title ILIKE ?)", like, like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron contar las entregas")
	}
	rows := make([]assessmentSubmissionRow, 0)
	if err := query.Order("s.submitted_at DESC, s.id DESC").Offset((page.Page - 1) * page.PerPage).Limit(page.PerPage).Scan(&rows).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar las entregas")
	}
	return mapAssessmentSubmissions(rows), total, nil
}

func (r *AssessmentRepository) ReviewSubmission(ctx context.Context, orgID, courseID int32, submissionID int64, reviewerID int32, score float64, feedback string) (*domainassessment.Submission, error) {
	db := connFromContext(ctx, r.db)
	var row models.AssessmentSubmission
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ? AND course_id = ?", submissionID, orgID, courseID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("entrega no encontrada")
		}
		return nil, apperrors.Internal("no se pudo cargar la entrega")
	}
	submission := domainassessment.Submission{ID: row.ID, OrganizationID: row.OrganizationID, CourseID: row.CourseID, LessonID: row.LessonID, StudentID: row.StudentID, Attempt: row.Attempt, ResponseText: row.ResponseText, ResponseURL: row.ResponseURL, Status: row.Status, Score: row.Score, Feedback: row.Feedback, SubmittedAt: row.SubmittedAt, ReviewedBy: row.ReviewedBy, ReviewedAt: row.ReviewedAt}
	if err := submission.Review(reviewerID, score, feedback, time.Now().UTC()); err != nil {
		return nil, err
	}
	if err := db.Model(&row).Updates(map[string]any{
		"status": submission.Status, "score": submission.Score, "feedback": submission.Feedback,
		"reviewed_by": submission.ReviewedBy, "reviewed_at": submission.ReviewedAt,
	}).Error; err != nil {
		return nil, apperrors.Internal("no se pudo guardar la revisión")
	}
	return r.submissionByID(db, orgID, courseID, submissionID)
}

type assessmentSubmissionRow struct {
	ID             int64      `gorm:"column:id"`
	OrganizationID int32      `gorm:"column:organization_id"`
	CourseID       int32      `gorm:"column:course_id"`
	LessonID       int32      `gorm:"column:lesson_id"`
	LessonTitle    string     `gorm:"column:lesson_title"`
	StudentID      int32      `gorm:"column:student_id"`
	StudentName    string     `gorm:"column:student_name"`
	StudentEmail   string     `gorm:"column:student_email"`
	Attempt        int        `gorm:"column:attempt_number"`
	ResponseText   string     `gorm:"column:response_text"`
	ResponseURL    *string    `gorm:"column:response_url"`
	Status         string     `gorm:"column:status"`
	Score          *float64   `gorm:"column:score"`
	Feedback       *string    `gorm:"column:feedback"`
	SubmittedAt    time.Time  `gorm:"column:submitted_at"`
	ReviewedBy     *int32     `gorm:"column:reviewed_by"`
	ReviewedAt     *time.Time `gorm:"column:reviewed_at"`
}

func submissionSelect(db *gorm.DB) *gorm.DB {
	return db.Table("assessment_submissions s").
		Select(`s.id, s.organization_id, s.course_id, s.lesson_id, l.title AS lesson_title,
			s.student_id, u.name AS student_name, u.email AS student_email,
			s.attempt_number, s.response_text, s.response_url, s.status, s.score,
			s.feedback, s.submitted_at, s.reviewed_by, s.reviewed_at`).
		Joins("JOIN lessons l ON l.id = s.lesson_id").
		Joins("JOIN users u ON u.id = s.student_id")
}

func (r *AssessmentRepository) submissionByID(db *gorm.DB, orgID, courseID int32, id int64) (*domainassessment.Submission, error) {
	var row assessmentSubmissionRow
	err := submissionSelect(db).Where("s.id = ? AND s.organization_id = ? AND s.course_id = ?", id, orgID, courseID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NotFound("entrega no encontrada")
	}
	if err != nil {
		return nil, apperrors.Internal("no se pudo cargar la entrega")
	}
	return mapAssessmentSubmission(row), nil
}

func mapAssessmentSubmissions(rows []assessmentSubmissionRow) []domainassessment.Submission {
	items := make([]domainassessment.Submission, 0, len(rows))
	for _, row := range rows {
		items = append(items, *mapAssessmentSubmission(row))
	}
	return items
}

func mapAssessmentSubmission(row assessmentSubmissionRow) *domainassessment.Submission {
	return &domainassessment.Submission{
		ID: row.ID, OrganizationID: row.OrganizationID, CourseID: row.CourseID,
		LessonID: row.LessonID, LessonTitle: row.LessonTitle, StudentID: row.StudentID,
		StudentName: row.StudentName, StudentEmail: row.StudentEmail, Attempt: row.Attempt,
		ResponseText: row.ResponseText, ResponseURL: row.ResponseURL, Status: row.Status,
		Score: row.Score, Feedback: row.Feedback, SubmittedAt: row.SubmittedAt,
		ReviewedBy: row.ReviewedBy, ReviewedAt: row.ReviewedAt,
	}
}
