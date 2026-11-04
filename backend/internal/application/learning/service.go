package learning

import (
	"context"
	"strings"
	"time"

	learningread "boilerplate-learner/internal/application/learning/read"
	appshared "boilerplate-learner/internal/application/shared"
	domaincohorts "boilerplate-learner/internal/domain/cohorts"
	"boilerplate-learner/internal/domain/curriculum"
	"boilerplate-learner/internal/domain/organizations"
	"boilerplate-learner/internal/shared/apperrors"
)

type Repository interface {
	CreateOrganization(context.Context, int32, string, string) (*organizations.Organization, error)
	ListOrganizations(context.Context, int32, PageQuery) ([]learningread.Organization, int64, error)
	UpdateOrganization(context.Context, int32, string, string) (*organizations.Organization, error)
	ArchiveOrganization(context.Context, int32) error
	RestoreOrganization(context.Context, int32) error
	OrganizationRole(context.Context, int32, int32) (string, error)
	OrganizationActive(context.Context, int32) (bool, error)
	AddMember(context.Context, int32, string, string) (*organizations.Membership, error)
	ListStudents(context.Context, int32, PageQuery) ([]learningread.Student, int64, error)
	RemoveMember(context.Context, int32, int32, int32) error
	CreateCourse(context.Context, int32, int32, string, string) (*curriculum.Course, error)
	UpdateCourse(context.Context, int32, int32, string, string) (*curriculum.Course, error)
	DeleteCourse(context.Context, int32, int32) error
	ListCourses(context.Context, int32, bool, int32, int32, PageQuery) ([]curriculum.Course, int64, error)
	CourseSummary(context.Context, int32, bool, int32, int32) (*CourseSummary, error)
	GetCourse(context.Context, int32, int32) (*curriculum.Course, error)
	PublishCourse(context.Context, int32, int32) (*curriculum.Course, error)
	AddModule(context.Context, int32, int32, string) (*curriculum.Module, error)
	UpdateModule(context.Context, int32, int32, int32, string) (*curriculum.Module, error)
	DeleteModule(context.Context, int32, int32, int32) error
	ReorderModules(context.Context, int32, int32, []int32) error
	AddLesson(context.Context, int32, int32, int32, string, string, string) (*curriculum.Lesson, error)
	UpdateLesson(context.Context, int32, int32, int32, int32, string, string, string) (*curriculum.Lesson, error)
	DeleteLesson(context.Context, int32, int32, int32, int32) error
	ReorderLessons(context.Context, int32, int32, int32, []int32) error
	ListCourseTrainers(context.Context, int32, int32) ([]learningread.CourseTrainer, error)
	AssignCourseTrainer(context.Context, int32, int32, int32) error
	RemoveCourseTrainer(context.Context, int32, int32, int32) error
	CourseTrainerAssigned(context.Context, int32, int32, int32) (bool, error)
	StudentHasCourseAccess(context.Context, int32, int32, int32) (bool, error)
	GetCourseProgress(context.Context, int32, int32, int32) (*learningread.CourseProgress, error)
	CompleteLesson(context.Context, int32, int32, int32, int32) (*learningread.LessonProgress, error)
	CreateCohort(context.Context, int32, int32, int32, string, *time.Time, *time.Time) (*learningread.Cohort, error)
	ListCohorts(context.Context, int32, int32, int32, string, PageQuery) ([]learningread.Cohort, int64, error)
	UpdateCohort(context.Context, int32, int32, string, *time.Time, *time.Time) (*learningread.Cohort, error)
	TransitionCohort(context.Context, int32, int32, string) error
	CohortInstructorAssigned(context.Context, int32, int32, int32) (bool, error)
	ListCohortInstructors(context.Context, int32, int32) ([]learningread.CohortInstructor, error)
	AssignCohortInstructor(context.Context, int32, int32, int32) error
	RemoveCohortInstructor(context.Context, int32, int32, int32) error
	ListCohortEnrollments(context.Context, int32, int32, PageQuery) ([]learningread.CohortEnrollment, int64, error)
	EnrollStudent(context.Context, int32, int32, int32, int32) (*learningread.CohortEnrollment, error)
	WithdrawStudent(context.Context, int32, int32, int32, int32, string) error
}

// Request models are owned by the application boundary so HTTP adapters do not
// define or bind anonymous transport-specific payloads.
type CreateOrganizationInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}
type UpdateOrganizationInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}
type AddMemberInput struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}
type AddStudentInput struct {
	Email string `json:"email"`
}
type CreateCourseInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
type UpdateCourseInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
type TitleInput struct {
	Title string `json:"title"`
}
type AddLessonInput struct {
	Title   string `json:"title"`
	Kind    string `json:"kind"`
	Content string `json:"content"`
}
type ReorderInput struct {
	IDs []int32 `json:"ids"`
}
type AssignTrainerInput struct {
	UserID int32 `json:"user_id"`
}
type CreateCohortInput struct {
	Name     string     `json:"name"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}
type UpdateCohortInput = CreateCohortInput
type TransitionCohortInput struct {
	Status string `json:"status"`
}
type AssignCohortInstructorInput struct {
	UserID int32 `json:"user_id"`
}
type EnrollStudentInput struct {
	UserID int32 `json:"user_id"`
}
type WithdrawStudentInput struct {
	Reason string `json:"reason"`
}

type Service struct {
	repo Repository
	tx   appshared.TxManager
}

func NewService(repo Repository, tx ...appshared.TxManager) *Service {
	service := &Service{repo: repo}
	if len(tx) > 0 {
		service.tx = tx[0]
	}
	return service
}

func (s *Service) CreateOrganization(ctx context.Context, userID int32, name, slug string) (*organizations.Organization, error) {
	name, slug = strings.TrimSpace(name), organizations.NormalizeSlug(slug)
	if name == "" || len([]rune(name)) > 160 || slug == "" || len(slug) > 100 {
		return nil, apperrors.Invalid("nombre y slug válidos son obligatorios")
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*organizations.Organization, error) {
		return s.repo.CreateOrganization(txctx, userID, name, slug)
	})
}

func (s *Service) ListOrganizations(ctx context.Context, userID int32, query PageQuery) (PageResult[learningread.Organization], error) {
	query = normalizePage(query)
	items, total, err := s.repo.ListOrganizations(ctx, userID, query)
	return pageResult(items, query, total), err
}

func (s *Service) UpdateOrganization(ctx context.Context, userID, organizationID int32, name, slug string) (*organizations.Organization, error) {
	name, slug = strings.TrimSpace(name), organizations.NormalizeSlug(slug)
	if name == "" || len([]rune(name)) > 160 || slug == "" || len(slug) > 100 {
		return nil, apperrors.Invalid("nombre y slug válidos son obligatorios")
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*organizations.Organization, error) {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return nil, err
		}
		return s.repo.UpdateOrganization(txctx, organizationID, name, slug)
	})
}

func (s *Service) ArchiveOrganization(ctx context.Context, userID, organizationID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return err
		}
		return s.repo.ArchiveOrganization(txctx, organizationID)
	})
}

func (s *Service) RestoreOrganization(ctx context.Context, userID, organizationID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		role, err := s.repo.OrganizationRole(txctx, userID, organizationID)
		if err != nil {
			return err
		}
		if role != organizations.RoleAdmin {
			return apperrors.Forbidden("no tienes permisos en esta organización")
		}
		return s.repo.RestoreOrganization(txctx, organizationID)
	})
}

func (s *Service) AddMember(ctx context.Context, actorID, organizationID int32, email, role string) (*organizations.Membership, error) {
	if err := organizations.ValidateRole(role); err != nil {
		return nil, err
	}
	if strings.TrimSpace(email) == "" {
		return nil, apperrors.Invalid("email obligatorio")
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*organizations.Membership, error) {
		if err := s.requireRole(txctx, actorID, organizationID, organizations.RoleAdmin); err != nil {
			return nil, err
		}
		return s.repo.AddMember(txctx, organizationID, strings.TrimSpace(email), role)
	})
}

func (s *Service) ListStudents(ctx context.Context, userID, organizationID int32, query PageQuery) (PageResult[learningread.Student], error) {
	query = normalizePage(query)
	return withinTx(ctx, s.tx, func(txctx context.Context) (PageResult[learningread.Student], error) {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin, organizations.RoleTrainer); err != nil {
			return PageResult[learningread.Student]{}, err
		}
		items, total, err := s.repo.ListStudents(txctx, organizationID, query)
		return pageResult(items, query, total), err
	})
}

func (s *Service) AddStudent(ctx context.Context, actorID, organizationID int32, email string) (*organizations.Membership, error) {
	return s.AddMember(ctx, actorID, organizationID, email, organizations.RoleStudent)
}

func (s *Service) RemoveStudent(ctx context.Context, actorID, organizationID, studentID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireRole(txctx, actorID, organizationID, organizations.RoleAdmin); err != nil {
			return err
		}
		return s.repo.RemoveMember(txctx, organizationID, studentID, actorID)
	})
}

func (s *Service) CreateCourse(ctx context.Context, userID, organizationID int32, title, description string) (*curriculum.Course, error) {
	if err := curriculum.ValidateCourse(title, description); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Course, error) {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin, organizations.RoleTrainer); err != nil {
			return nil, err
		}
		return s.repo.CreateCourse(txctx, userID, organizationID, strings.TrimSpace(title), strings.TrimSpace(description))
	})
}

func (s *Service) UpdateCourse(ctx context.Context, userID, organizationID, courseID int32, title, description string) (*curriculum.Course, error) {
	if err := curriculum.ValidateCourse(title, description); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Course, error) {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return nil, err
		}
		return s.repo.UpdateCourse(txctx, organizationID, courseID, strings.TrimSpace(title), strings.TrimSpace(description))
	})
}

func (s *Service) DeleteCourse(ctx context.Context, userID, organizationID, courseID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return err
		}
		return s.repo.DeleteCourse(txctx, organizationID, courseID)
	})
}

func (s *Service) ListCourses(ctx context.Context, userID, organizationID int32, query PageQuery) (PageResult[curriculum.Course], error) {
	query = normalizePage(query)
	return withinTx(ctx, s.tx, func(txctx context.Context) (PageResult[curriculum.Course], error) {
		role, err := s.requireActiveMembership(txctx, userID, organizationID)
		if err != nil {
			return PageResult[curriculum.Course]{}, err
		}
		trainerID := int32(0)
		studentID := int32(0)
		if role == organizations.RoleTrainer {
			trainerID = userID
		} else if role == organizations.RoleStudent {
			studentID = userID
		}
		courses, total, err := s.repo.ListCourses(txctx, organizationID, role == organizations.RoleStudent, trainerID, studentID, query)
		if err != nil {
			return PageResult[curriculum.Course]{}, err
		}
		result := pageResult(courses, query, total)
		result.Summary, err = s.repo.CourseSummary(txctx, organizationID, role == organizations.RoleStudent, trainerID, studentID)
		if err != nil {
			return PageResult[curriculum.Course]{}, err
		}
		return result, nil
	})
}

func (s *Service) GetCourse(ctx context.Context, userID, organizationID, courseID int32) (*curriculum.Course, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Course, error) {
		role, err := s.requireActiveMembership(txctx, userID, organizationID)
		if err != nil {
			return nil, err
		}
		course, err := s.repo.GetCourse(txctx, organizationID, courseID)
		if err != nil {
			return nil, err
		}
		if role == organizations.RoleStudent {
			if course.Status != curriculum.CoursePublished {
				return nil, apperrors.NotFound("curso no encontrado")
			}
			allowed, err := s.repo.StudentHasCourseAccess(txctx, organizationID, courseID, userID)
			if err != nil {
				return nil, err
			}
			if !allowed {
				return nil, apperrors.NotFound("curso no encontrado")
			}
			progress, err := s.repo.GetCourseProgress(txctx, organizationID, courseID, userID)
			if err != nil {
				return nil, err
			}
			unlocked := make(map[int32]bool, len(progress.Lessons))
			for _, lesson := range progress.Lessons {
				unlocked[lesson.LessonID] = lesson.Unlocked
			}
			for moduleIndex := range course.Modules {
				for lessonIndex := range course.Modules[moduleIndex].Lessons {
					lesson := &course.Modules[moduleIndex].Lessons[lessonIndex]
					if !unlocked[lesson.ID] {
						lesson.Content = ""
					}
				}
			}
		}
		if role == organizations.RoleTrainer {
			assigned, err := s.repo.CourseTrainerAssigned(txctx, organizationID, courseID, userID)
			if err != nil {
				return nil, err
			}
			if !assigned {
				return nil, apperrors.NotFound("curso no encontrado")
			}
		}
		return course, nil
	})
}

func (s *Service) GetCourseProgress(ctx context.Context, userID, organizationID, courseID int32) (*learningread.CourseProgress, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) (*learningread.CourseProgress, error) {
		role, err := s.requireActiveMembership(txctx, userID, organizationID)
		if err != nil {
			return nil, err
		}
		if role != organizations.RoleStudent {
			return nil, apperrors.Forbidden("el progreso personal solo está disponible para estudiantes")
		}
		allowed, err := s.repo.StudentHasCourseAccess(txctx, organizationID, courseID, userID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, apperrors.NotFound("curso no encontrado")
		}
		return s.repo.GetCourseProgress(txctx, organizationID, courseID, userID)
	})
}

func (s *Service) CompleteLesson(ctx context.Context, userID, organizationID, courseID, lessonID int32) (*learningread.LessonProgress, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) (*learningread.LessonProgress, error) {
		role, err := s.requireActiveMembership(txctx, userID, organizationID)
		if err != nil {
			return nil, err
		}
		if role != organizations.RoleStudent {
			return nil, apperrors.Forbidden("solo los estudiantes pueden completar lecciones")
		}
		allowed, err := s.repo.StudentHasCourseAccess(txctx, organizationID, courseID, userID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, apperrors.NotFound("curso no encontrado")
		}
		return s.repo.CompleteLesson(txctx, organizationID, courseID, lessonID, userID)
	})
}

func (s *Service) PublishCourse(ctx context.Context, userID, organizationID, courseID int32) (*curriculum.Course, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Course, error) {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return nil, err
		}
		return s.repo.PublishCourse(txctx, organizationID, courseID)
	})
}

func (s *Service) AddModule(ctx context.Context, userID, organizationID, courseID int32, title string) (*curriculum.Module, error) {
	if err := curriculum.ValidateModule(title); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Module, error) {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return nil, err
		}
		return s.repo.AddModule(txctx, organizationID, courseID, strings.TrimSpace(title))
	})
}

func (s *Service) UpdateModule(ctx context.Context, userID, organizationID, courseID, moduleID int32, title string) (*curriculum.Module, error) {
	if err := curriculum.ValidateModule(title); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Module, error) {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return nil, err
		}
		return s.repo.UpdateModule(txctx, organizationID, courseID, moduleID, strings.TrimSpace(title))
	})
}

func (s *Service) DeleteModule(ctx context.Context, userID, organizationID, courseID, moduleID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return err
		}
		return s.repo.DeleteModule(txctx, organizationID, courseID, moduleID)
	})
}

func (s *Service) ReorderModules(ctx context.Context, userID, organizationID, courseID int32, ids []int32) error {
	if err := validateOrder(ids); err != nil {
		return err
	}
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return err
		}
		return s.repo.ReorderModules(txctx, organizationID, courseID, ids)
	})
}

func (s *Service) AddLesson(ctx context.Context, userID, organizationID, courseID, moduleID int32, title, kind, content string) (*curriculum.Lesson, error) {
	if err := curriculum.ValidateLesson(title, kind); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Lesson, error) {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return nil, err
		}
		return s.repo.AddLesson(txctx, organizationID, courseID, moduleID, strings.TrimSpace(title), kind, content)
	})
}

func (s *Service) UpdateLesson(ctx context.Context, userID, organizationID, courseID, moduleID, lessonID int32, title, kind, content string) (*curriculum.Lesson, error) {
	if err := curriculum.ValidateLesson(title, kind); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*curriculum.Lesson, error) {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return nil, err
		}
		return s.repo.UpdateLesson(txctx, organizationID, courseID, moduleID, lessonID, strings.TrimSpace(title), kind, content)
	})
}

func (s *Service) DeleteLesson(ctx context.Context, userID, organizationID, courseID, moduleID, lessonID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return err
		}
		return s.repo.DeleteLesson(txctx, organizationID, courseID, moduleID, lessonID)
	})
}

func (s *Service) ReorderLessons(ctx context.Context, userID, organizationID, courseID, moduleID int32, ids []int32) error {
	if err := validateOrder(ids); err != nil {
		return err
	}
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return err
		}
		return s.repo.ReorderLessons(txctx, organizationID, courseID, moduleID, ids)
	})
}

func (s *Service) ListCourseTrainers(ctx context.Context, userID, organizationID, courseID int32) ([]learningread.CourseTrainer, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) ([]learningread.CourseTrainer, error) {
		if err := s.requireCourseManager(txctx, userID, organizationID, courseID); err != nil {
			return nil, err
		}
		return s.repo.ListCourseTrainers(txctx, organizationID, courseID)
	})
}

func (s *Service) AssignCourseTrainer(ctx context.Context, userID, organizationID, courseID, trainerID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return err
		}
		return s.repo.AssignCourseTrainer(txctx, organizationID, courseID, trainerID)
	})
}

func (s *Service) RemoveCourseTrainer(ctx context.Context, userID, organizationID, courseID, trainerID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return err
		}
		return s.repo.RemoveCourseTrainer(txctx, organizationID, courseID, trainerID)
	})
}

func (s *Service) CreateCohort(ctx context.Context, userID, organizationID, courseID int32, input CreateCohortInput) (*learningread.Cohort, error) {
	if err := domaincohorts.Validate(input.Name, input.StartsAt, input.EndsAt); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*learningread.Cohort, error) {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return nil, err
		}
		return s.repo.CreateCohort(txctx, userID, organizationID, courseID, strings.TrimSpace(input.Name), input.StartsAt, input.EndsAt)
	})
}

func (s *Service) ListCohorts(ctx context.Context, userID, organizationID, courseID int32, query PageQuery) (PageResult[learningread.Cohort], error) {
	query = normalizePage(query)
	return withinTx(ctx, s.tx, func(txctx context.Context) (PageResult[learningread.Cohort], error) {
		role, err := s.requireActiveMembership(txctx, userID, organizationID)
		if err != nil {
			return PageResult[learningread.Cohort]{}, err
		}
		items, total, err := s.repo.ListCohorts(txctx, organizationID, courseID, userID, role, query)
		return pageResult(items, query, total), err
	})
}

func (s *Service) UpdateCohort(ctx context.Context, userID, organizationID, cohortID int32, input UpdateCohortInput) (*learningread.Cohort, error) {
	if err := domaincohorts.Validate(input.Name, input.StartsAt, input.EndsAt); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*learningread.Cohort, error) {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return nil, err
		}
		return s.repo.UpdateCohort(txctx, organizationID, cohortID, strings.TrimSpace(input.Name), input.StartsAt, input.EndsAt)
	})
}

func (s *Service) TransitionCohort(ctx context.Context, userID, organizationID, cohortID int32, status string) error {
	if status != domaincohorts.StatusActive && status != domaincohorts.StatusCompleted && status != domaincohorts.StatusCancelled {
		return apperrors.Invalid("estado de cohorte no válido")
	}
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return err
		}
		return s.repo.TransitionCohort(txctx, organizationID, cohortID, status)
	})
}

func (s *Service) ListCohortInstructors(ctx context.Context, userID, organizationID, cohortID int32) ([]learningread.CohortInstructor, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) ([]learningread.CohortInstructor, error) {
		if err := s.requireCohortManager(txctx, userID, organizationID, cohortID); err != nil {
			return nil, err
		}
		return s.repo.ListCohortInstructors(txctx, organizationID, cohortID)
	})
}

func (s *Service) AssignCohortInstructor(ctx context.Context, userID, organizationID, cohortID, trainerID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return err
		}
		return s.repo.AssignCohortInstructor(txctx, organizationID, cohortID, trainerID)
	})
}

func (s *Service) RemoveCohortInstructor(ctx context.Context, userID, organizationID, cohortID, trainerID int32) error {
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireRole(txctx, userID, organizationID, organizations.RoleAdmin); err != nil {
			return err
		}
		return s.repo.RemoveCohortInstructor(txctx, organizationID, cohortID, trainerID)
	})
}

func (s *Service) ListCohortEnrollments(ctx context.Context, userID, organizationID, cohortID int32, query PageQuery) (PageResult[learningread.CohortEnrollment], error) {
	query = normalizePage(query)
	return withinTx(ctx, s.tx, func(txctx context.Context) (PageResult[learningread.CohortEnrollment], error) {
		role, err := s.requireActiveMembership(txctx, userID, organizationID)
		if err != nil {
			return PageResult[learningread.CohortEnrollment]{}, err
		}
		if role == organizations.RoleTrainer {
			assigned, err := s.repo.CohortInstructorAssigned(txctx, organizationID, cohortID, userID)
			if err != nil {
				return PageResult[learningread.CohortEnrollment]{}, err
			}
			if !assigned {
				return PageResult[learningread.CohortEnrollment]{}, apperrors.NotFound("cohorte no encontrada")
			}
		} else if role == organizations.RoleStudent {
			return PageResult[learningread.CohortEnrollment]{}, apperrors.Forbidden("no tienes permisos para consultar matrículas")
		}
		items, total, err := s.repo.ListCohortEnrollments(txctx, organizationID, cohortID, query)
		return pageResult(items, query, total), err
	})
}

func (s *Service) EnrollStudent(ctx context.Context, userID, organizationID, cohortID, studentID int32) (*learningread.CohortEnrollment, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) (*learningread.CohortEnrollment, error) {
		if err := s.requireCohortManager(txctx, userID, organizationID, cohortID); err != nil {
			return nil, err
		}
		return s.repo.EnrollStudent(txctx, organizationID, cohortID, studentID, userID)
	})
}

func (s *Service) WithdrawStudent(ctx context.Context, userID, organizationID, cohortID, studentID int32, reason string) error {
	if len([]rune(reason)) > 500 {
		return apperrors.Invalid("el motivo de baja no puede superar 500 caracteres")
	}
	return withinTxErr(ctx, s.tx, func(txctx context.Context) error {
		if err := s.requireCohortManager(txctx, userID, organizationID, cohortID); err != nil {
			return err
		}
		return s.repo.WithdrawStudent(txctx, organizationID, cohortID, studentID, userID, strings.TrimSpace(reason))
	})
}

func validateOrder(ids []int32) error {
	if len(ids) == 0 {
		return apperrors.Invalid("el orden debe incluir elementos")
	}
	seen := make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		if id < 1 {
			return apperrors.Invalid("identificador de elemento inválido")
		}
		if _, exists := seen[id]; exists {
			return apperrors.Invalid("el orden contiene elementos duplicados")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func (s *Service) requireRole(ctx context.Context, userID, organizationID int32, allowed ...string) error {
	active, err := s.repo.OrganizationActive(ctx, organizationID)
	if err != nil {
		return err
	}
	if !active {
		return apperrors.NotFound("organización no encontrada")
	}
	role, err := s.repo.OrganizationRole(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if err := validateMemberRole(role); err != nil {
		return err
	}
	for _, candidate := range allowed {
		if candidate == role {
			return nil
		}
	}
	return apperrors.Forbidden("no tienes permisos en esta organización")
}

func (s *Service) requireCourseManager(ctx context.Context, userID, organizationID, courseID int32) error {
	role, err := s.requireActiveMembership(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if role == organizations.RoleAdmin {
		return nil
	}
	if role != organizations.RoleTrainer {
		return apperrors.Forbidden("no tienes permisos para gestionar este curso")
	}
	assigned, err := s.repo.CourseTrainerAssigned(ctx, organizationID, courseID, userID)
	if err != nil {
		return err
	}
	if !assigned {
		return apperrors.NotFound("curso no encontrado")
	}
	return nil
}

func (s *Service) requireCohortManager(ctx context.Context, userID, organizationID, cohortID int32) error {
	role, err := s.requireActiveMembership(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if role == organizations.RoleAdmin {
		return nil
	}
	if role != organizations.RoleTrainer {
		return apperrors.Forbidden("no tienes permisos para gestionar esta cohorte")
	}
	assigned, err := s.repo.CohortInstructorAssigned(ctx, organizationID, cohortID, userID)
	if err != nil {
		return err
	}
	if !assigned {
		return apperrors.NotFound("cohorte no encontrada")
	}
	return nil
}

func (s *Service) requireActiveMembership(ctx context.Context, userID, organizationID int32) (string, error) {
	active, err := s.repo.OrganizationActive(ctx, organizationID)
	if err != nil {
		return "", err
	}
	if !active {
		return "", apperrors.NotFound("organización no encontrada")
	}
	role, err := s.repo.OrganizationRole(ctx, userID, organizationID)
	if err != nil {
		return "", err
	}
	if err := validateMemberRole(role); err != nil {
		return "", err
	}
	return role, nil
}

func validateMemberRole(role string) error {
	if err := organizations.ValidateRole(role); err != nil {
		return apperrors.Forbidden("rol de organización no válido")
	}
	return nil
}

func withinTx[T any](ctx context.Context, tx appshared.TxManager, action func(context.Context) (T, error)) (T, error) {
	var result T
	if tx == nil {
		return action(ctx)
	}
	err := tx.WithinTx(ctx, func(txctx context.Context) error {
		var err error
		result, err = action(txctx)
		return err
	})
	return result, err
}

func withinTxErr(ctx context.Context, tx appshared.TxManager, action func(context.Context) error) error {
	if tx == nil {
		return action(ctx)
	}
	return tx.WithinTx(ctx, action)
}
