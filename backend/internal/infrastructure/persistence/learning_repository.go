package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	learningapp "boilerplate-learner/internal/application/learning"
	learningread "boilerplate-learner/internal/application/learning/read"
	domaincohorts "boilerplate-learner/internal/domain/cohorts"
	"boilerplate-learner/internal/domain/curriculum"
	"boilerplate-learner/internal/domain/organizations"
	"boilerplate-learner/internal/infrastructure/persistence/models"
	"boilerplate-learner/internal/shared/apperrors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LearningRepository struct{ db *gorm.DB }

func NewLearningRepository(db *gorm.DB) *LearningRepository { return &LearningRepository{db: db} }

func (r *LearningRepository) CreateOrganization(ctx context.Context, userID int32, name, slug string) (*organizations.Organization, error) {
	db := connFromContext(ctx, r.db)
	org := models.Organization{Name: name, Slug: slug, CreatedBy: userID}
	if err := db.Create(&org).Error; err != nil {
		return nil, mapLearningDBError(err, "slug de organización ya existe")
	}
	if err := db.Create(&models.OrganizationMembership{OrganizationID: org.ID, UserID: userID, Role: organizations.RoleAdmin, Status: "active"}).Error; err != nil {
		return nil, mapLearningDBError(err, "no se pudo crear la membresía de administrador")
	}
	return &organizations.Organization{ID: org.ID, Name: org.Name, Slug: org.Slug}, nil
}

func (r *LearningRepository) ListOrganizations(ctx context.Context, userID int32, page learningapp.PageQuery) ([]learningread.Organization, int64, error) {
	db := connFromContext(ctx, r.db)
	items := make([]learningread.Organization, 0)
	query := db.Table("organizations").
		Select(`organizations.id, organizations.name, organizations.slug, m.role,
			organizations.archived_at,
			(SELECT COUNT(*) FROM courses c WHERE c.organization_id = organizations.id) AS course_count,
			(SELECT COUNT(*) FROM organization_memberships sm WHERE sm.organization_id = organizations.id AND sm.status = 'active' AND sm.role = ?) AS student_count`, organizations.RoleStudent).
		Joins("JOIN organization_memberships m ON m.organization_id = organizations.id").
		Where("m.user_id = ? AND m.status = ?", userID, "active")
	if page.Search != "" {
		query = query.Where("(organizations.name ILIKE ? OR organizations.slug ILIKE ?)", "%"+page.Search+"%", "%"+page.Search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar las organizaciones")
	}
	err := query.Order("organizations.archived_at NULLS FIRST, organizations.name").Offset((page.Page - 1) * page.PerPage).Limit(page.PerPage).Scan(&items).Error
	return items, total, err
}

func (r *LearningRepository) UpdateOrganization(ctx context.Context, orgID int32, name, slug string) (*organizations.Organization, error) {
	db := connFromContext(ctx, r.db)
	var row models.Organization
	if err := db.First(&row, "id = ?", orgID).Error; err != nil {
		return nil, mapLearningDBError(err, "organización no encontrada")
	}
	if err := db.Model(&row).Updates(map[string]any{"name": name, "slug": slug}).Error; err != nil {
		return nil, mapLearningDBError(err, "el slug de organización ya existe")
	}
	row.Name, row.Slug = name, slug
	return &organizations.Organization{ID: row.ID, Name: row.Name, Slug: row.Slug}, nil
}

func (r *LearningRepository) ArchiveOrganization(ctx context.Context, orgID int32) error {
	db := connFromContext(ctx, r.db)
	result := db.Model(&models.Organization{}).Where("id = ? AND archived_at IS NULL", orgID).Update("archived_at", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return apperrors.Internal("no se pudo archivar la organización")
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("organización activa no encontrada")
	}
	return nil
}

func (r *LearningRepository) RestoreOrganization(ctx context.Context, orgID int32) error {
	result := connFromContext(ctx, r.db).Model(&models.Organization{}).Where("id = ? AND archived_at IS NOT NULL", orgID).Update("archived_at", nil)
	if result.Error != nil {
		return apperrors.Internal("no se pudo restaurar la organización")
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("organización archivada no encontrada")
	}
	return nil
}

func (r *LearningRepository) OrganizationActive(ctx context.Context, orgID int32) (bool, error) {
	var row models.Organization
	err := connFromContext(ctx, r.db).Select("id", "archived_at").First(&row, "id = ?", orgID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, apperrors.NotFound("organización no encontrada")
	}
	if err != nil {
		return false, apperrors.Internal("no se pudo consultar la organización")
	}
	return row.ArchivedAt == nil, nil
}

func (r *LearningRepository) AddMember(ctx context.Context, orgID int32, email, role string) (*organizations.Membership, error) {
	db := connFromContext(ctx, r.db)
	var result *organizations.Membership
	err := db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("LOWER(email) = LOWER(?)", email).First(&user).Error; err != nil {
			return mapLearningDBError(err, "no existe un usuario con ese email")
		}
		var m models.OrganizationMembership
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("organization_id = ? AND user_id = ?", orgID, user.ID).First(&m).Error
		if err == nil {
			if m.Status == "active" {
				return apperrors.Conflict("el usuario ya pertenece a esta organización")
			}
			if err := tx.Model(&m).Updates(map[string]any{"role": role, "status": "active"}).Error; err != nil {
				return mapLearningDBError(err, "no se pudo reactivar la membresía")
			}
			m.Role, m.Status = role, "active"
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			m = models.OrganizationMembership{OrganizationID: orgID, UserID: user.ID, Role: role, Status: "active"}
			if err := tx.Create(&m).Error; err != nil {
				return mapLearningDBError(err, "el usuario ya pertenece a esta organización")
			}
		} else {
			return apperrors.Internal("no se pudo consultar la membresía")
		}
		result = &organizations.Membership{OrganizationID: m.OrganizationID, UserID: m.UserID, Role: m.Role}
		return nil
	})
	return result, err
}

func (r *LearningRepository) ListStudents(ctx context.Context, orgID int32, page learningapp.PageQuery) ([]learningread.Student, int64, error) {
	db := connFromContext(ctx, r.db)
	students := make([]learningread.Student, 0)
	query := db.Table("organization_memberships AS memberships").
		Select("users.id AS user_id, users.name AS name, users.email AS email, memberships.role AS role").
		Joins("JOIN users ON users.id = memberships.user_id").
		Where("memberships.organization_id = ? AND memberships.status = ? AND memberships.role = ?", orgID, "active", organizations.RoleStudent)
	if page.Search != "" {
		query = query.Where("(users.name ILIKE ? OR users.email ILIKE ?)", "%"+page.Search+"%", "%"+page.Search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar los estudiantes")
	}
	if err := query.Order("users.name ASC").Offset((page.Page - 1) * page.PerPage).Limit(page.PerPage).Scan(&students).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar los estudiantes")
	}
	return students, total, nil
}

func (r *LearningRepository) RemoveMember(ctx context.Context, orgID, userID, actorID int32) error {
	db := connFromContext(ctx, r.db)
	result := db.Model(&models.OrganizationMembership{}).
		Where("organization_id = ? AND user_id = ? AND role = ? AND status = 'active'", orgID, userID, organizations.RoleStudent).
		Update("status", "inactive")
	if result.Error != nil {
		return apperrors.Internal("no se pudo eliminar el estudiante")
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("estudiante no encontrado")
	}
	now := time.Now()
	actor := actorID
	if err := db.Model(&models.CohortEnrollment{}).
		Where("user_id = ? AND status = 'active' AND cohort_id IN (SELECT id FROM cohorts WHERE organization_id = ?)", userID, orgID).
		Updates(map[string]any{"status": "withdrawn", "withdrawn_by": &actor, "withdrawn_at": &now, "withdrawal_reason": "membership revoked"}).Error; err != nil {
		return apperrors.Internal("no se pudieron cerrar las matrículas activas del estudiante")
	}
	return nil
}

func (r *LearningRepository) CreateCourse(ctx context.Context, userID, orgID int32, title, description string) (*curriculum.Course, error) {
	db := connFromContext(ctx, r.db)
	slug := organizations.NormalizeSlug(title)
	if slug == "" {
		return nil, apperrors.Invalid("título de curso inválido")
	}
	var count int64
	if err := db.Model(&models.Course{}).Where("organization_id = ? AND slug LIKE ?", orgID, slug+"%").Count(&count).Error; err != nil {
		return nil, apperrors.Internal("no se pudo crear el curso")
	}
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, count+1)
	}
	row := models.Course{OrganizationID: orgID, Title: title, Slug: slug, Description: description, Status: curriculum.CourseDraft, CreatedBy: userID}
	if err := db.Create(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "no se pudo crear el curso")
	}
	role, err := r.OrganizationRole(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if role == organizations.RoleTrainer {
		if err := db.Create(&models.CourseTrainer{CourseID: row.ID, TrainerID: userID}).Error; err != nil {
			return nil, apperrors.Internal("no se pudo asignar el curso a su docente creador")
		}
	}
	return mapCourse(&row), nil
}

func (r *LearningRepository) UpdateCourse(ctx context.Context, orgID, courseID int32, title, description string) (*curriculum.Course, error) {
	db := connFromContext(ctx, r.db)
	var row models.Course
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ?", courseID, orgID).First(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "curso no encontrado")
	}
	if err := (curriculum.Course{Status: row.Status}).EnsureDraft(); err != nil {
		return nil, err
	}
	if err := db.Model(&row).Updates(map[string]any{"title": title, "description": description}).Error; err != nil {
		return nil, mapLearningDBError(err, "no se pudo actualizar el curso")
	}
	row.Title, row.Description = title, description
	return mapCourse(&row), nil
}

func (r *LearningRepository) DeleteCourse(ctx context.Context, orgID, courseID int32) error {
	db := connFromContext(ctx, r.db)
	var row models.Course
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ?", courseID, orgID).First(&row).Error; err != nil {
		return mapLearningDBError(err, "curso no encontrado")
	}
	if err := (curriculum.Course{Status: row.Status}).EnsureDraft(); err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM lessons WHERE module_id IN (SELECT id FROM course_modules WHERE course_id = ?)", courseID).Error; err != nil {
		return apperrors.Internal("no se pudo eliminar el contenido del curso")
	}
	if err := db.Where("course_id = ?", courseID).Delete(&models.CourseModule{}).Error; err != nil {
		return apperrors.Internal("no se pudieron eliminar los módulos del curso")
	}
	if err := db.Delete(&row).Error; err != nil {
		return apperrors.Internal("no se pudo eliminar el curso")
	}
	return nil
}

func (r *LearningRepository) ListCourses(ctx context.Context, orgID int32, publishedOnly bool, trainerID, studentID int32, page learningapp.PageQuery) ([]curriculum.Course, int64, error) {
	db := connFromContext(ctx, r.db)
	var rows []models.Course
	query := db.Where("organization_id = ?", orgID)
	if publishedOnly {
		query = query.Where("status = ?", curriculum.CoursePublished)
	}
	if trainerID > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM course_trainers ct WHERE ct.course_id = courses.id AND ct.trainer_id = ?)", trainerID)
	}
	if studentID > 0 {
		query = query.Where(`EXISTS (
			SELECT 1 FROM cohorts c
			JOIN cohort_enrollments ce ON ce.cohort_id = c.id
			WHERE c.course_id = courses.id AND c.organization_id = courses.organization_id
			AND c.status IN ? AND ce.user_id = ? AND ce.status = 'active'
		)`, []string{domaincohorts.StatusActive, domaincohorts.StatusCompleted}, studentID)
	}
	if page.Search != "" {
		query = query.Where("(title ILIKE ? OR description ILIKE ?)", "%"+page.Search+"%", "%"+page.Search+"%")
	}
	var total int64
	if err := query.Model(&models.Course{}).Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar los cursos")
	}
	if err := query.Preload("Modules", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).Order("created_at DESC").Offset((page.Page - 1) * page.PerPage).Limit(page.PerPage).Find(&rows).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar los cursos")
	}
	out := make([]curriculum.Course, 0, len(rows))
	for i := range rows {
		out = append(out, *mapCourse(&rows[i]))
	}
	return out, total, nil
}

func (r *LearningRepository) CourseSummary(ctx context.Context, orgID int32, publishedOnly bool, trainerID, studentID int32) (*learningapp.CourseSummary, error) {
	query := connFromContext(ctx, r.db).Table("courses c").
		Joins("LEFT JOIN course_modules cm ON cm.course_id = c.id").
		Where("c.organization_id = ?", orgID)
	if publishedOnly {
		query = query.Where("c.status = ?", curriculum.CoursePublished)
	}
	if trainerID > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM course_trainers ct WHERE ct.course_id = c.id AND ct.trainer_id = ?)", trainerID)
	}
	if studentID > 0 {
		query = query.Where(`EXISTS (
			SELECT 1 FROM cohorts cohort
			JOIN cohort_enrollments ce ON ce.cohort_id = cohort.id
			WHERE cohort.course_id = c.id AND cohort.organization_id = c.organization_id
			AND cohort.status IN ? AND ce.user_id = ? AND ce.status = 'active'
		)`, []string{domaincohorts.StatusActive, domaincohorts.StatusCompleted}, studentID)
	}
	var summary learningapp.CourseSummary
	err := query.Select(`COUNT(DISTINCT c.id) AS course_count,
		COUNT(DISTINCT c.id) FILTER (WHERE c.status = ?) AS published_count,
		COUNT(cm.id) AS module_count`, curriculum.CoursePublished).Scan(&summary).Error
	if err != nil {
		return nil, apperrors.Internal("no se pudo calcular el resumen de cursos")
	}
	return &summary, nil
}

func (r *LearningRepository) ListCourseTrainers(ctx context.Context, orgID, courseID int32) ([]learningread.CourseTrainer, error) {
	db := connFromContext(ctx, r.db)
	var courseCount int64
	if err := db.Model(&models.Course{}).Where("id = ? AND organization_id = ?", courseID, orgID).Count(&courseCount).Error; err != nil {
		return nil, apperrors.Internal("no se pudo validar el curso")
	}
	if courseCount == 0 {
		return nil, apperrors.NotFound("curso no encontrado")
	}
	items := make([]learningread.CourseTrainer, 0)
	err := db.Table("organization_memberships m").
		Select(`users.id AS user_id, users.name, users.email,
			EXISTS (SELECT 1 FROM course_trainers ct WHERE ct.course_id = ? AND ct.trainer_id = users.id) AS assigned`, courseID).
		Joins("JOIN users ON users.id = m.user_id").
		Where("m.organization_id = ? AND m.status = 'active' AND m.role = ?", orgID, organizations.RoleTrainer).
		Order("users.name ASC").Scan(&items).Error
	if err != nil {
		return nil, apperrors.Internal("no se pudieron cargar los docentes del curso")
	}
	return items, nil
}

func (r *LearningRepository) AssignCourseTrainer(ctx context.Context, orgID, courseID, trainerID int32) error {
	db := connFromContext(ctx, r.db)
	var courseCount int64
	if err := db.Model(&models.Course{}).Where("id = ? AND organization_id = ?", courseID, orgID).Count(&courseCount).Error; err != nil {
		return apperrors.Internal("no se pudo validar el curso")
	}
	if courseCount == 0 {
		return apperrors.NotFound("curso no encontrado")
	}
	var membership models.OrganizationMembership
	if err := db.Where("organization_id = ? AND user_id = ? AND status = 'active' AND role = ?", orgID, trainerID, organizations.RoleTrainer).First(&membership).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.Invalid("el usuario no es un docente activo de la organización")
		}
		return apperrors.Internal("no se pudo validar el docente")
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.CourseTrainer{CourseID: courseID, TrainerID: trainerID}).Error; err != nil {
		return apperrors.Internal("no se pudo asignar el docente al curso")
	}
	return nil
}

func (r *LearningRepository) RemoveCourseTrainer(ctx context.Context, orgID, courseID, trainerID int32) error {
	db := connFromContext(ctx, r.db)
	result := db.Where("course_id = ? AND trainer_id = ? AND course_id IN (SELECT id FROM courses WHERE organization_id = ?)", courseID, trainerID, orgID).Delete(&models.CourseTrainer{})
	if result.Error != nil {
		return apperrors.Internal("no se pudo quitar el docente del curso")
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("asignación de docente no encontrada")
	}
	return nil
}

func (r *LearningRepository) CourseTrainerAssigned(ctx context.Context, orgID, courseID, trainerID int32) (bool, error) {
	var count int64
	err := connFromContext(ctx, r.db).Table("course_trainers ct").Joins("JOIN courses c ON c.id = ct.course_id").
		Where("ct.course_id = ? AND ct.trainer_id = ? AND c.organization_id = ?", courseID, trainerID, orgID).Count(&count).Error
	if err != nil {
		return false, apperrors.Internal("no se pudo comprobar la asignación del docente")
	}
	return count > 0, nil
}

func (r *LearningRepository) StudentHasCourseAccess(ctx context.Context, orgID, courseID, studentID int32) (bool, error) {
	var count int64
	err := connFromContext(ctx, r.db).Table("cohorts c").
		Joins("JOIN cohort_enrollments ce ON ce.cohort_id = c.id").
		Joins("JOIN courses ON courses.id = c.course_id").
		Where("c.organization_id = ? AND c.course_id = ? AND courses.status = ? AND c.status IN ? AND ce.user_id = ? AND ce.status = 'active'",
			orgID, courseID, curriculum.CoursePublished, []string{domaincohorts.StatusActive, domaincohorts.StatusCompleted}, studentID).
		Count(&count).Error
	if err != nil {
		return false, apperrors.Internal("no se pudo comprobar la inscripción al curso")
	}
	return count > 0, nil
}

func (r *LearningRepository) GetCourseProgress(ctx context.Context, orgID, courseID, studentID int32) (*learningread.CourseProgress, error) {
	db := connFromContext(ctx, r.db)
	var courseCount int64
	if err := db.Model(&models.Course{}).Where("id = ? AND organization_id = ? AND status = ?", courseID, orgID, curriculum.CoursePublished).Count(&courseCount).Error; err != nil {
		return nil, apperrors.Internal("no se pudo validar el curso")
	}
	if courseCount == 0 {
		return nil, apperrors.NotFound("curso no encontrado")
	}
	progress := &learningread.CourseProgress{CourseID: courseID, Lessons: make([]learningread.LessonProgress, 0)}
	err := db.Table("lessons").
		Select("lessons.id AS lesson_id, student_lesson_progress.completed_at").
		Joins("JOIN course_modules ON course_modules.id = lessons.module_id").
		Joins("LEFT JOIN student_lesson_progress ON student_lesson_progress.lesson_id = lessons.id AND student_lesson_progress.user_id = ?", studentID).
		Where("course_modules.course_id = ?", courseID).
		Order("course_modules.position ASC, lessons.position ASC").Scan(&progress.Lessons).Error
	if err != nil {
		return nil, apperrors.Internal("no se pudo cargar el progreso del curso")
	}
	progress.TotalLessons = int64(len(progress.Lessons))
	previousLessonsCompleted := true
	for index := range progress.Lessons {
		lesson := &progress.Lessons[index]
		lesson.Unlocked = previousLessonsCompleted
		// A completed lesson keeps the next lesson unlocked; a pending lesson stops
		// unlocking further content.
		if lesson.CompletedAt != nil {
			progress.CompletedLessons++
		} else {
			previousLessonsCompleted = false
		}
	}
	if progress.TotalLessons > 0 {
		progress.Percent = int(progress.CompletedLessons * 100 / progress.TotalLessons)
	}
	return progress, nil
}

func (r *LearningRepository) CompleteLesson(ctx context.Context, orgID, courseID, lessonID, studentID int32) (*learningread.LessonProgress, error) {
	db := connFromContext(ctx, r.db)
	type lessonPosition struct {
		ModulePosition int    `gorm:"column:module_position"`
		Position       int    `gorm:"column:lesson_position"`
		Kind           string `gorm:"column:kind"`
	}
	var target lessonPosition
	err := db.Table("lessons").
		Joins("JOIN course_modules ON course_modules.id = lessons.module_id").
		Joins("JOIN courses ON courses.id = course_modules.course_id").
		Select("course_modules.position AS module_position, lessons.position AS lesson_position, lessons.kind").
		Where("lessons.id = ? AND courses.id = ? AND courses.organization_id = ? AND courses.status = ?", lessonID, courseID, orgID, curriculum.CoursePublished).
		Take(&target).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("lección no encontrada")
		}
		return nil, apperrors.Internal("no se pudo validar la lección")
	}
	var pendingPrevious int64
	err = db.Table("lessons previous_lesson").
		Joins("JOIN course_modules previous_module ON previous_module.id = previous_lesson.module_id").
		Where(`previous_module.course_id = ? AND (
			previous_module.position < ? OR
			(previous_module.position = ? AND previous_lesson.position < ?)
		) AND NOT EXISTS (
			SELECT 1 FROM student_lesson_progress progress
			WHERE progress.lesson_id = previous_lesson.id AND progress.user_id = ?
		)`, courseID, target.ModulePosition, target.ModulePosition, target.Position, studentID).
		Count(&pendingPrevious).Error
	if err != nil {
		return nil, apperrors.Internal("no se pudo validar el orden de las lecciones")
	}
	if pendingPrevious > 0 {
		return nil, apperrors.Conflict("completa las lecciones anteriores para desbloquear esta lección")
	}
	if target.Kind == "task" {
		var submissionCount int64
		if err := db.Model(&models.AssessmentSubmission{}).Where("student_id = ? AND lesson_id = ?", studentID, lessonID).Count(&submissionCount).Error; err != nil {
			return nil, apperrors.Internal("no se pudo validar la entrega de la actividad")
		}
		if submissionCount == 0 {
			return nil, apperrors.Conflict("envía una entrega antes de completar esta actividad")
		}
	}
	row := models.StudentLessonProgress{UserID: studentID, LessonID: lessonID}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "lesson_id"}}, DoNothing: true}).Create(&row).Error; err != nil {
		return nil, apperrors.Internal("no se pudo guardar el progreso de la lección")
	}
	var completed models.StudentLessonProgress
	if err := db.Where("user_id = ? AND lesson_id = ?", studentID, lessonID).Take(&completed).Error; err != nil {
		return nil, apperrors.Internal("no se pudo cargar el progreso de la lección")
	}
	return &learningread.LessonProgress{LessonID: completed.LessonID, CompletedAt: &completed.CompletedAt}, nil
}

func (r *LearningRepository) CreateCohort(ctx context.Context, userID, orgID, courseID int32, name string, startsAt, endsAt *time.Time) (*learningread.Cohort, error) {
	db := connFromContext(ctx, r.db)
	var course models.Course
	if err := db.Where("id = ? AND organization_id = ? AND status = ?", courseID, orgID, curriculum.CoursePublished).First(&course).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Invalid("solo se pueden crear cohortes para cursos publicados")
		}
		return nil, apperrors.Internal("no se pudo validar el curso")
	}
	row := models.Cohort{OrganizationID: orgID, CourseID: courseID, Name: name, StartsAt: startsAt, EndsAt: endsAt, Status: domaincohorts.StatusDraft, CreatedBy: userID}
	if err := db.Create(&row).Error; err != nil {
		return nil, apperrors.Internal("no se pudo crear la cohorte")
	}
	return cohortView(db, row.ID, orgID)
}

func (r *LearningRepository) ListCohorts(ctx context.Context, orgID, courseID, userID int32, role string, page learningapp.PageQuery) ([]learningread.Cohort, int64, error) {
	db := connFromContext(ctx, r.db)
	query := db.Table("cohorts c").
		Select(`c.id, c.organization_id, c.course_id, courses.title AS course_title, c.name,
		c.starts_at, c.ends_at, c.status,
		(SELECT COUNT(*) FROM cohort_enrollments ce WHERE ce.cohort_id = c.id AND ce.status = 'active') AS enrollment_count`).
		Joins("JOIN courses ON courses.id = c.course_id").
		Where("c.organization_id = ?", orgID)
	if courseID > 0 {
		query = query.Where("c.course_id = ?", courseID)
	}
	switch role {
	case organizations.RoleTrainer:
		query = query.Where("EXISTS (SELECT 1 FROM cohort_instructors ci WHERE ci.cohort_id = c.id AND ci.trainer_id = ?)", userID)
	case organizations.RoleStudent:
		query = query.Where("c.status IN ? AND EXISTS (SELECT 1 FROM cohort_enrollments ce WHERE ce.cohort_id = c.id AND ce.user_id = ? AND ce.status = 'active')", []string{domaincohorts.StatusActive, domaincohorts.StatusCompleted}, userID)
	}
	if page.Search != "" {
		query = query.Where("(c.name ILIKE ? OR courses.title ILIKE ?)", "%"+page.Search+"%", "%"+page.Search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar las cohortes")
	}
	items := make([]learningread.Cohort, 0)
	if err := query.Order("c.starts_at DESC NULLS LAST, c.id DESC").Offset((page.Page - 1) * page.PerPage).Limit(page.PerPage).Scan(&items).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar las cohortes")
	}
	return items, total, nil
}

func (r *LearningRepository) UpdateCohort(ctx context.Context, orgID, cohortID int32, name string, startsAt, endsAt *time.Time) (*learningread.Cohort, error) {
	db := connFromContext(ctx, r.db)
	var row models.Cohort
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ?", cohortID, orgID).First(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "cohorte no encontrada")
	}
	if row.Status != domaincohorts.StatusDraft {
		return nil, apperrors.Conflict("solo se pueden editar cohortes en borrador")
	}
	if err := db.Model(&row).Updates(map[string]any{"name": name, "starts_at": startsAt, "ends_at": endsAt}).Error; err != nil {
		return nil, apperrors.Internal("no se pudo actualizar la cohorte")
	}
	return cohortView(db, cohortID, orgID)
}

func (r *LearningRepository) TransitionCohort(ctx context.Context, orgID, cohortID int32, status string) error {
	db := connFromContext(ctx, r.db)
	var row models.Cohort
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ?", cohortID, orgID).First(&row).Error; err != nil {
		return mapLearningDBError(err, "cohorte no encontrada")
	}
	if !domaincohorts.CanTransition(row.Status, status) {
		return apperrors.Conflict("transición de estado de cohorte no permitida")
	}
	if err := db.Model(&row).Update("status", status).Error; err != nil {
		return apperrors.Internal("no se pudo cambiar el estado de la cohorte")
	}
	return nil
}

func (r *LearningRepository) CohortInstructorAssigned(ctx context.Context, orgID, cohortID, trainerID int32) (bool, error) {
	var count int64
	err := connFromContext(ctx, r.db).Table("cohort_instructors ci").Joins("JOIN cohorts c ON c.id = ci.cohort_id").
		Where("c.organization_id = ? AND ci.cohort_id = ? AND ci.trainer_id = ?", orgID, cohortID, trainerID).Count(&count).Error
	if err != nil {
		return false, apperrors.Internal("no se pudo comprobar la asignación del docente")
	}
	return count > 0, nil
}

func (r *LearningRepository) ListCohortInstructors(ctx context.Context, orgID, cohortID int32) ([]learningread.CohortInstructor, error) {
	db := connFromContext(ctx, r.db)
	if _, err := cohortView(db, cohortID, orgID); err != nil {
		return nil, err
	}
	items := make([]learningread.CohortInstructor, 0)
	err := db.Table("organization_memberships m").
		Select(`users.id AS user_id, users.name, users.email,
		EXISTS (SELECT 1 FROM cohort_instructors ci WHERE ci.cohort_id = ? AND ci.trainer_id = users.id) AS assigned`, cohortID).
		Joins("JOIN users ON users.id = m.user_id").
		Where("m.organization_id = ? AND m.status = 'active' AND m.role = ?", orgID, organizations.RoleTrainer).
		Order("users.name ASC").Scan(&items).Error
	if err != nil {
		return nil, apperrors.Internal("no se pudieron cargar los docentes de la cohorte")
	}
	return items, nil
}

func (r *LearningRepository) AssignCohortInstructor(ctx context.Context, orgID, cohortID, trainerID int32) error {
	db := connFromContext(ctx, r.db)
	if _, err := cohortView(db, cohortID, orgID); err != nil {
		return err
	}
	var membership models.OrganizationMembership
	if err := db.Where("organization_id = ? AND user_id = ? AND status = 'active' AND role = ?", orgID, trainerID, organizations.RoleTrainer).First(&membership).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.Invalid("el usuario no es docente activo de la organización")
		}
		return apperrors.Internal("no se pudo validar el docente")
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.CohortInstructor{CohortID: cohortID, TrainerID: trainerID}).Error; err != nil {
		return apperrors.Internal("no se pudo asignar el docente a la cohorte")
	}
	return nil
}

func (r *LearningRepository) RemoveCohortInstructor(ctx context.Context, orgID, cohortID, trainerID int32) error {
	db := connFromContext(ctx, r.db)
	result := db.Where("cohort_id = ? AND trainer_id = ? AND cohort_id IN (SELECT id FROM cohorts WHERE organization_id = ?)", cohortID, trainerID, orgID).Delete(&models.CohortInstructor{})
	if result.Error != nil {
		return apperrors.Internal("no se pudo quitar el docente de la cohorte")
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("asignación de docente no encontrada")
	}
	return nil
}

func (r *LearningRepository) ListCohortEnrollments(ctx context.Context, orgID, cohortID int32, page learningapp.PageQuery) ([]learningread.CohortEnrollment, int64, error) {
	db := connFromContext(ctx, r.db)
	query := db.Table("cohort_enrollments ce").
		Select("ce.id, ce.cohort_id, ce.user_id, users.name, users.email, ce.status, ce.enrolled_at, ce.withdrawn_at, ce.withdrawal_reason").
		Joins("JOIN users ON users.id = ce.user_id").
		Joins("JOIN cohorts c ON c.id = ce.cohort_id").
		Where("c.organization_id = ? AND ce.cohort_id = ?", orgID, cohortID)
	if page.Search != "" {
		query = query.Where("(users.name ILIKE ? OR users.email ILIKE ?)", "%"+page.Search+"%", "%"+page.Search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar las matrículas")
	}
	items := make([]learningread.CohortEnrollment, 0)
	if err := query.Order("ce.enrolled_at DESC, ce.id DESC").Offset((page.Page - 1) * page.PerPage).Limit(page.PerPage).Scan(&items).Error; err != nil {
		return nil, 0, apperrors.Internal("no se pudieron cargar las matrículas")
	}
	return items, total, nil
}

func (r *LearningRepository) EnrollStudent(ctx context.Context, orgID, cohortID, studentID, enrolledBy int32) (*learningread.CohortEnrollment, error) {
	db := connFromContext(ctx, r.db)
	var cohort models.Cohort
	if err := db.Where("id = ? AND organization_id = ? AND status = ?", cohortID, orgID, domaincohorts.StatusActive).First(&cohort).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Conflict("solo se puede matricular en cohortes activas")
		}
		return nil, apperrors.Internal("no se pudo validar la cohorte")
	}
	var membership models.OrganizationMembership
	if err := db.Where("organization_id = ? AND user_id = ? AND status = 'active' AND role = ?", orgID, studentID, organizations.RoleStudent).First(&membership).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Invalid("el usuario no es estudiante activo de la organización")
		}
		return nil, apperrors.Internal("no se pudo validar el estudiante")
	}
	row := models.CohortEnrollment{CohortID: cohortID, UserID: studentID, Status: domaincohorts.EnrollmentActive, EnrolledBy: enrolledBy}
	if err := db.Create(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "el estudiante ya tiene una matrícula activa en esta cohorte")
	}
	return enrollmentView(db, row.ID, orgID)
}

func (r *LearningRepository) WithdrawStudent(ctx context.Context, orgID, cohortID, studentID, withdrawnBy int32, reason string) error {
	db := connFromContext(ctx, r.db)
	now := time.Now()
	withdrawnByValue := withdrawnBy
	updates := map[string]any{"status": domaincohorts.EnrollmentWithdrawn, "withdrawn_by": &withdrawnByValue, "withdrawn_at": &now, "withdrawal_reason": nullableString(reason)}
	result := db.Model(&models.CohortEnrollment{}).
		Where("cohort_id = ? AND user_id = ? AND status = 'active' AND cohort_id IN (SELECT id FROM cohorts WHERE organization_id = ?)", cohortID, studentID, orgID).
		Updates(updates)
	if result.Error != nil {
		return apperrors.Internal("no se pudo dar de baja al estudiante")
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("matrícula activa no encontrada")
	}
	return nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func cohortView(db *gorm.DB, cohortID, orgID int32) (*learningread.Cohort, error) {
	var item learningread.Cohort
	err := db.Table("cohorts c").
		Select(`c.id, c.organization_id, c.course_id, courses.title AS course_title, c.name,
		c.starts_at, c.ends_at, c.status,
		(SELECT COUNT(*) FROM cohort_enrollments ce WHERE ce.cohort_id = c.id AND ce.status = 'active') AS enrollment_count`).
		Joins("JOIN courses ON courses.id = c.course_id").Where("c.id = ? AND c.organization_id = ?", cohortID, orgID).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NotFound("cohorte no encontrada")
	}
	if err != nil {
		return nil, apperrors.Internal("no se pudo cargar la cohorte")
	}
	return &item, nil
}

func enrollmentView(db *gorm.DB, enrollmentID int64, orgID int32) (*learningread.CohortEnrollment, error) {
	var item learningread.CohortEnrollment
	err := db.Table("cohort_enrollments ce").
		Select("ce.id, ce.cohort_id, ce.user_id, users.name, users.email, ce.status, ce.enrolled_at, ce.withdrawn_at, ce.withdrawal_reason").
		Joins("JOIN users ON users.id = ce.user_id").Joins("JOIN cohorts c ON c.id = ce.cohort_id").
		Where("ce.id = ? AND c.organization_id = ?", enrollmentID, orgID).Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NotFound("matrícula no encontrada")
	}
	if err != nil {
		return nil, apperrors.Internal("no se pudo cargar la matrícula")
	}
	return &item, nil
}

func (r *LearningRepository) GetCourse(ctx context.Context, orgID, courseID int32) (*curriculum.Course, error) {
	db := connFromContext(ctx, r.db)
	var row models.Course
	query := db.Preload("Modules", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).Preload("Modules.Lessons", func(db *gorm.DB) *gorm.DB { return db.Order("position ASC") }).Where("id = ? AND organization_id = ?", courseID, orgID)
	if err := query.First(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "curso no encontrado")
	}
	return mapCourse(&row), nil
}

func (r *LearningRepository) PublishCourse(ctx context.Context, orgID, courseID int32) (*curriculum.Course, error) {
	db := connFromContext(ctx, r.db)
	var row models.Course
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ?", courseID, orgID).First(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "curso no encontrado")
	}
	var lessonCount int64
	if err := db.Model(&models.Lesson{}).
		Joins("JOIN course_modules ON course_modules.id = lessons.module_id").
		Where("course_modules.course_id = ?", courseID).Count(&lessonCount).Error; err != nil {
		return nil, apperrors.Internal("no se pudo validar el contenido del curso")
	}
	domainCourse := curriculum.Course{ID: row.ID, OrganizationID: row.OrganizationID, Title: row.Title, Slug: row.Slug, Description: row.Description, Status: row.Status}
	if err := domainCourse.Publish(int(lessonCount)); err != nil {
		return nil, err
	}
	if err := db.Model(&row).Update("status", curriculum.CoursePublished).Error; err != nil {
		return nil, apperrors.Internal("no se pudo publicar el curso")
	}
	row.Status = curriculum.CoursePublished
	return mapCourse(&row), nil
}

func (r *LearningRepository) AddModule(ctx context.Context, orgID, courseID int32, title string) (*curriculum.Module, error) {
	db := connFromContext(ctx, r.db)
	var course models.Course
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ?", courseID, orgID).First(&course).Error; err != nil {
		return nil, mapLearningDBError(err, "curso no encontrado")
	}
	domainCourse := curriculum.Course{Status: course.Status}
	if err := domainCourse.EnsureDraft(); err != nil {
		return nil, err
	}
	var position int
	if err := db.Model(&models.CourseModule{}).Where("course_id = ?", courseID).Select("COALESCE(MAX(position), 0)").Scan(&position).Error; err != nil {
		return nil, apperrors.Internal("no se pudo ordenar el módulo")
	}
	row := models.CourseModule{CourseID: courseID, Title: title, Position: position + 1}
	if err := db.Create(&row).Error; err != nil {
		return nil, apperrors.Internal("no se pudo crear el módulo")
	}
	return &curriculum.Module{ID: row.ID, CourseID: row.CourseID, Title: row.Title, Position: row.Position}, nil
}

func (r *LearningRepository) UpdateModule(ctx context.Context, orgID, courseID, moduleID int32, title string) (*curriculum.Module, error) {
	db := connFromContext(ctx, r.db)
	if err := lockDraftCourse(db, orgID, courseID); err != nil {
		return nil, err
	}
	var row models.CourseModule
	if err := db.Where("id = ? AND course_id = ?", moduleID, courseID).First(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "módulo no encontrado")
	}
	if err := db.Model(&row).Update("title", title).Error; err != nil {
		return nil, apperrors.Internal("no se pudo actualizar el módulo")
	}
	row.Title = title
	return &curriculum.Module{ID: row.ID, CourseID: row.CourseID, Title: row.Title, Position: row.Position}, nil
}

func (r *LearningRepository) DeleteModule(ctx context.Context, orgID, courseID, moduleID int32) error {
	db := connFromContext(ctx, r.db)
	if err := lockDraftCourse(db, orgID, courseID); err != nil {
		return err
	}
	var module models.CourseModule
	if err := db.Where("id = ? AND course_id = ?", moduleID, courseID).First(&module).Error; err != nil {
		return mapLearningDBError(err, "módulo no encontrado")
	}
	if err := db.Where("module_id = ?", moduleID).Delete(&models.Lesson{}).Error; err != nil {
		return apperrors.Internal("no se pudieron eliminar las lecciones")
	}
	if err := db.Delete(&module).Error; err != nil {
		return apperrors.Internal("no se pudo eliminar el módulo")
	}
	return nil
}

func (r *LearningRepository) ReorderModules(ctx context.Context, orgID, courseID int32, ids []int32) error {
	db := connFromContext(ctx, r.db)
	if err := lockDraftCourse(db, orgID, courseID); err != nil {
		return err
	}
	var modules []models.CourseModule
	if err := db.Where("course_id = ?", courseID).Find(&modules).Error; err != nil {
		return apperrors.Internal("no se pudo cargar el orden de módulos")
	}
	if len(modules) != len(ids) {
		return apperrors.Invalid("el orden debe contener todos los módulos del curso")
	}
	valid := make(map[int32]struct{}, len(modules))
	for _, module := range modules {
		valid[module.ID] = struct{}{}
	}
	for _, id := range ids {
		if _, ok := valid[id]; !ok {
			return apperrors.Invalid("el orden contiene un módulo ajeno al curso")
		}
	}
	if err := db.Model(&models.CourseModule{}).Where("course_id = ?", courseID).Update("position", gorm.Expr("-id")).Error; err != nil {
		return apperrors.Internal("no se pudo preparar el orden de módulos")
	}
	for position, id := range ids {
		if err := db.Model(&models.CourseModule{}).Where("id = ? AND course_id = ?", id, courseID).Update("position", position+1).Error; err != nil {
			return apperrors.Internal("no se pudo actualizar el orden de módulos")
		}
	}
	return nil
}

func (r *LearningRepository) AddLesson(ctx context.Context, orgID, courseID, moduleID int32, title, kind, content string) (*curriculum.Lesson, error) {
	db := connFromContext(ctx, r.db)
	var module models.CourseModule
	if err := db.Joins("JOIN courses ON courses.id = course_modules.course_id").Where("course_modules.id = ? AND course_modules.course_id = ? AND courses.organization_id = ?", moduleID, courseID, orgID).First(&module).Error; err != nil {
		return nil, mapLearningDBError(err, "módulo no encontrado")
	}
	var course models.Course
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Select("status").First(&course, "id = ?", courseID).Error; err != nil {
		return nil, mapLearningDBError(err, "curso no encontrado")
	}
	if err := (curriculum.Course{Status: course.Status}).EnsureDraft(); err != nil {
		return nil, err
	}
	var position int
	if err := db.Model(&models.Lesson{}).Where("module_id = ?", moduleID).Select("COALESCE(MAX(position), 0)").Scan(&position).Error; err != nil {
		return nil, apperrors.Internal("no se pudo ordenar la lección")
	}
	row := models.Lesson{ModuleID: moduleID, Title: title, Kind: kind, Content: content, Position: position + 1}
	if err := db.Create(&row).Error; err != nil {
		return nil, apperrors.Internal("no se pudo crear la lección")
	}
	return &curriculum.Lesson{ID: row.ID, ModuleID: row.ModuleID, Title: row.Title, Kind: row.Kind, Content: row.Content, Position: row.Position}, nil
}

func (r *LearningRepository) UpdateLesson(ctx context.Context, orgID, courseID, moduleID, lessonID int32, title, kind, content string) (*curriculum.Lesson, error) {
	db := connFromContext(ctx, r.db)
	if err := lockDraftCourse(db, orgID, courseID); err != nil {
		return nil, err
	}
	var row models.Lesson
	if err := db.Joins("JOIN course_modules ON course_modules.id = lessons.module_id").Where("lessons.id = ? AND lessons.module_id = ? AND course_modules.course_id = ?", lessonID, moduleID, courseID).First(&row).Error; err != nil {
		return nil, mapLearningDBError(err, "lección no encontrada")
	}
	if err := db.Model(&row).Updates(map[string]any{"title": title, "kind": kind, "content": content}).Error; err != nil {
		return nil, apperrors.Internal("no se pudo actualizar la lección")
	}
	row.Title, row.Kind, row.Content = title, kind, content
	return &curriculum.Lesson{ID: row.ID, ModuleID: row.ModuleID, Title: row.Title, Kind: row.Kind, Content: row.Content, Position: row.Position}, nil
}

func (r *LearningRepository) DeleteLesson(ctx context.Context, orgID, courseID, moduleID, lessonID int32) error {
	db := connFromContext(ctx, r.db)
	if err := lockDraftCourse(db, orgID, courseID); err != nil {
		return err
	}
	result := db.Where(`id = ? AND module_id = ? AND module_id IN
		(SELECT id FROM course_modules WHERE course_id = ?)`, lessonID, moduleID, courseID).Delete(&models.Lesson{})
	if result.Error != nil {
		return apperrors.Internal("no se pudo eliminar la lección")
	}
	if result.RowsAffected == 0 {
		return apperrors.NotFound("lección no encontrada")
	}
	return nil
}

func (r *LearningRepository) ReorderLessons(ctx context.Context, orgID, courseID, moduleID int32, ids []int32) error {
	db := connFromContext(ctx, r.db)
	if err := lockDraftCourse(db, orgID, courseID); err != nil {
		return err
	}
	var module models.CourseModule
	if err := db.Where("id = ? AND course_id = ?", moduleID, courseID).First(&module).Error; err != nil {
		return mapLearningDBError(err, "módulo no encontrado")
	}
	var lessons []models.Lesson
	if err := db.Where("module_id = ?", moduleID).Find(&lessons).Error; err != nil {
		return apperrors.Internal("no se pudo cargar el orden de lecciones")
	}
	if len(lessons) != len(ids) {
		return apperrors.Invalid("el orden debe contener todas las lecciones del módulo")
	}
	valid := make(map[int32]struct{}, len(lessons))
	for _, lesson := range lessons {
		valid[lesson.ID] = struct{}{}
	}
	for _, id := range ids {
		if _, ok := valid[id]; !ok {
			return apperrors.Invalid("el orden contiene una lección ajena al módulo")
		}
	}
	if err := db.Model(&models.Lesson{}).Where("module_id = ?", moduleID).Update("position", gorm.Expr("-id")).Error; err != nil {
		return apperrors.Internal("no se pudo preparar el orden de lecciones")
	}
	for position, id := range ids {
		if err := db.Model(&models.Lesson{}).Where("id = ? AND module_id = ?", id, moduleID).Update("position", position+1).Error; err != nil {
			return apperrors.Internal("no se pudo actualizar el orden de lecciones")
		}
	}
	return nil
}

func lockDraftCourse(db *gorm.DB, orgID, courseID int32) error {
	var course models.Course
	if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND organization_id = ?", courseID, orgID).First(&course).Error; err != nil {
		return mapLearningDBError(err, "curso no encontrado")
	}
	return (curriculum.Course{Status: course.Status}).EnsureDraft()
}

func (r *LearningRepository) OrganizationRole(ctx context.Context, userID, orgID int32) (string, error) {
	db := connFromContext(ctx, r.db)
	var membership models.OrganizationMembership
	if err := db.Clauses(clause.Locking{Strength: "SHARE"}).Where("organization_id = ? AND user_id = ? AND status = ?", orgID, userID, "active").First(&membership).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.Forbidden("no tienes acceso a esta organización")
		}
		return "", apperrors.Internal("no se pudo comprobar la membresía")
	}
	return membership.Role, nil
}

func mapCourse(row *models.Course) *curriculum.Course {
	course := &curriculum.Course{ID: row.ID, OrganizationID: row.OrganizationID, Title: row.Title, Slug: row.Slug, Description: row.Description, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
	for _, m := range row.Modules {
		module := curriculum.Module{ID: m.ID, CourseID: m.CourseID, Title: m.Title, Position: m.Position}
		for _, l := range m.Lessons {
			module.Lessons = append(module.Lessons, curriculum.Lesson{ID: l.ID, ModuleID: l.ModuleID, Title: l.Title, Kind: l.Kind, Content: l.Content, Position: l.Position})
		}
		course.Modules = append(course.Modules, module)
	}
	return course
}

func mapLearningDBError(err error, conflictMessage string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.NotFound(conflictMessage)
	}
	if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		return apperrors.Conflict(conflictMessage)
	}
	return apperrors.Internal("error de persistencia")
}
