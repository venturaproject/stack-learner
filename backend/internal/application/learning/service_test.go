package learning

import (
	"context"
	"testing"
	"time"

	learningread "boilerplate-learner/internal/application/learning/read"
	"boilerplate-learner/internal/domain/curriculum"
	"boilerplate-learner/internal/domain/organizations"
	"boilerplate-learner/internal/shared/apperrors"
)

type learningRepoStub struct {
	role            string
	active          bool
	trainerAssigned bool
	courses         []curriculum.Course
}

func (r *learningRepoStub) CreateOrganization(context.Context, int32, string, string) (*organizations.Organization, error) {
	return nil, nil
}
func (r *learningRepoStub) ListOrganizations(context.Context, int32, PageQuery) ([]learningread.Organization, int64, error) {
	return nil, 0, nil
}
func (r *learningRepoStub) UpdateOrganization(context.Context, int32, string, string) (*organizations.Organization, error) {
	return nil, nil
}
func (r *learningRepoStub) ArchiveOrganization(context.Context, int32) error { return nil }
func (r *learningRepoStub) RestoreOrganization(context.Context, int32) error { return nil }
func (r *learningRepoStub) OrganizationActive(context.Context, int32) (bool, error) {
	return r.active, nil
}
func (r *learningRepoStub) OrganizationRole(context.Context, int32, int32) (string, error) {
	return r.role, nil
}
func (r *learningRepoStub) AddMember(context.Context, int32, string, string) (*organizations.Membership, error) {
	return nil, nil
}
func (r *learningRepoStub) ListStudents(context.Context, int32, PageQuery) ([]learningread.Student, int64, error) {
	return nil, 0, nil
}
func (r *learningRepoStub) RemoveMember(context.Context, int32, int32, int32) error { return nil }
func (r *learningRepoStub) CreateCourse(context.Context, int32, int32, string, string) (*curriculum.Course, error) {
	return nil, nil
}
func (r *learningRepoStub) UpdateCourse(context.Context, int32, int32, string, string) (*curriculum.Course, error) {
	return nil, nil
}
func (r *learningRepoStub) DeleteCourse(context.Context, int32, int32) error { return nil }
func (r *learningRepoStub) ListCourses(_ context.Context, _ int32, publishedOnly bool, _ int32, _ int32, _ PageQuery) ([]curriculum.Course, int64, error) {
	if !publishedOnly {
		return r.courses, int64(len(r.courses)), nil
	}
	items := make([]curriculum.Course, 0, len(r.courses))
	for _, course := range r.courses {
		if course.Status == curriculum.CoursePublished {
			items = append(items, course)
		}
	}
	return items, int64(len(items)), nil
}
func (r *learningRepoStub) CourseSummary(context.Context, int32, bool, int32, int32) (*CourseSummary, error) {
	return &CourseSummary{}, nil
}
func (r *learningRepoStub) GetCourse(_ context.Context, _, _ int32) (*curriculum.Course, error) {
	return &r.courses[0], nil
}
func (r *learningRepoStub) PublishCourse(context.Context, int32, int32) (*curriculum.Course, error) {
	return nil, nil
}
func (r *learningRepoStub) AddModule(context.Context, int32, int32, string) (*curriculum.Module, error) {
	return nil, nil
}
func (r *learningRepoStub) UpdateModule(context.Context, int32, int32, int32, string) (*curriculum.Module, error) {
	return nil, nil
}
func (r *learningRepoStub) DeleteModule(context.Context, int32, int32, int32) error     { return nil }
func (r *learningRepoStub) ReorderModules(context.Context, int32, int32, []int32) error { return nil }
func (r *learningRepoStub) AddLesson(context.Context, int32, int32, int32, string, string, string) (*curriculum.Lesson, error) {
	return nil, nil
}
func (r *learningRepoStub) UpdateLesson(context.Context, int32, int32, int32, int32, string, string, string) (*curriculum.Lesson, error) {
	return nil, nil
}
func (r *learningRepoStub) DeleteLesson(context.Context, int32, int32, int32, int32) error {
	return nil
}
func (r *learningRepoStub) ReorderLessons(context.Context, int32, int32, int32, []int32) error {
	return nil
}
func (r *learningRepoStub) ListCourseTrainers(context.Context, int32, int32) ([]learningread.CourseTrainer, error) {
	return nil, nil
}
func (r *learningRepoStub) AssignCourseTrainer(context.Context, int32, int32, int32) error {
	return nil
}
func (r *learningRepoStub) RemoveCourseTrainer(context.Context, int32, int32, int32) error {
	return nil
}
func (r *learningRepoStub) CourseTrainerAssigned(context.Context, int32, int32, int32) (bool, error) {
	return r.trainerAssigned, nil
}
func (r *learningRepoStub) StudentHasCourseAccess(context.Context, int32, int32, int32) (bool, error) {
	return true, nil
}
func (r *learningRepoStub) GetCourseProgress(context.Context, int32, int32, int32) (*learningread.CourseProgress, error) {
	return &learningread.CourseProgress{}, nil
}
func (r *learningRepoStub) CompleteLesson(context.Context, int32, int32, int32, int32) (*learningread.LessonProgress, error) {
	return &learningread.LessonProgress{}, nil
}
func (r *learningRepoStub) CreateCohort(context.Context, int32, int32, int32, string, *time.Time, *time.Time) (*learningread.Cohort, error) {
	return nil, nil
}
func (r *learningRepoStub) ListCohorts(context.Context, int32, int32, int32, string, PageQuery) ([]learningread.Cohort, int64, error) {
	return nil, 0, nil
}
func (r *learningRepoStub) UpdateCohort(context.Context, int32, int32, string, *time.Time, *time.Time) (*learningread.Cohort, error) {
	return nil, nil
}
func (r *learningRepoStub) TransitionCohort(context.Context, int32, int32, string) error { return nil }
func (r *learningRepoStub) CohortInstructorAssigned(context.Context, int32, int32, int32) (bool, error) {
	return r.trainerAssigned, nil
}
func (r *learningRepoStub) ListCohortInstructors(context.Context, int32, int32) ([]learningread.CohortInstructor, error) {
	return nil, nil
}
func (r *learningRepoStub) AssignCohortInstructor(context.Context, int32, int32, int32) error {
	return nil
}
func (r *learningRepoStub) RemoveCohortInstructor(context.Context, int32, int32, int32) error {
	return nil
}
func (r *learningRepoStub) ListCohortEnrollments(context.Context, int32, int32, PageQuery) ([]learningread.CohortEnrollment, int64, error) {
	return nil, 0, nil
}
func (r *learningRepoStub) EnrollStudent(context.Context, int32, int32, int32, int32) (*learningread.CohortEnrollment, error) {
	return nil, nil
}
func (r *learningRepoStub) WithdrawStudent(context.Context, int32, int32, int32, int32, string) error {
	return nil
}

func TestListCoursesOnlyShowsPublishedCoursesToStudents(t *testing.T) {
	repo := &learningRepoStub{role: organizations.RoleStudent, active: true, courses: []curriculum.Course{
		{ID: 1, Status: curriculum.CourseDraft},
		{ID: 2, Status: curriculum.CoursePublished},
	}}
	got, err := NewService(repo).ListCourses(context.Background(), 12, 34, PageQuery{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].ID != 2 {
		t.Fatalf("student course list = %+v, want only published course", got)
	}
}

func TestStudentCannotReadDraftCourse(t *testing.T) {
	repo := &learningRepoStub{role: organizations.RoleStudent, active: true, courses: []curriculum.Course{{ID: 1, Status: curriculum.CourseDraft}}}
	_, err := NewService(repo).GetCourse(context.Background(), 12, 34, 1)
	if apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("draft course should be hidden, got %v", err)
	}
}

func TestProgressIsStudentSelfService(t *testing.T) {
	repo := &learningRepoStub{role: organizations.RoleStudent, active: true}
	service := NewService(repo)
	if _, err := service.GetCourseProgress(context.Background(), 10, 20, 30); err != nil {
		t.Fatalf("student should read own course progress: %v", err)
	}
	if _, err := service.CompleteLesson(context.Background(), 10, 20, 30, 40); err != nil {
		t.Fatalf("student should complete an accessible lesson: %v", err)
	}

	repo.role = organizations.RoleAdmin
	if _, err := service.GetCourseProgress(context.Background(), 10, 20, 30); apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("administrator must not use student progress endpoint, got %v", err)
	}
	if _, err := service.CompleteLesson(context.Background(), 10, 20, 30, 40); apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("administrator must not complete a lesson for a student, got %v", err)
	}
}

func TestOnlyOrganizationAdminCanAddMembers(t *testing.T) {
	repo := &learningRepoStub{role: organizations.RoleTrainer, active: true}
	_, err := NewService(repo).AddMember(context.Background(), 12, 34, "learner@example.com", organizations.RoleStudent)
	if apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("trainer add-member should be forbidden, got %v", err)
	}
}

func TestStudentCannotListOrganizationStudents(t *testing.T) {
	repo := &learningRepoStub{role: organizations.RoleStudent, active: true}
	_, err := NewService(repo).ListStudents(context.Background(), 12, 34, PageQuery{Page: 1, PerPage: 25})
	if apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("student list-members should be forbidden, got %v", err)
	}
}

func TestArchivedOrganizationCannotBeUsed(t *testing.T) {
	repo := &learningRepoStub{role: organizations.RoleAdmin, active: false}
	if _, err := NewService(repo).ListCourses(context.Background(), 7, 9, PageQuery{Page: 1, PerPage: 25}); apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("archived organization should be unavailable, got %v", err)
	}
}

func TestTrainerCannotManageUnassignedCourse(t *testing.T) {
	repo := &learningRepoStub{role: organizations.RoleTrainer, active: true}
	_, err := NewService(repo).UpdateCourse(context.Background(), 7, 9, 11, "Title", "")
	if apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("unassigned course should be hidden from trainer, got %v", err)
	}
}
