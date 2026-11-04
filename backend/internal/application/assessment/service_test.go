package assessment

import (
	"context"
	"testing"
	"time"

	domainassessment "boilerplate-learner/internal/domain/assessment"
)

type assessmentRepoStub struct {
	role     string
	assigned bool
}

func (r *assessmentRepoStub) ActiveOrganizationRole(context.Context, int32, int32) (string, error) {
	return r.role, nil
}
func (*assessmentRepoStub) ValidateStudentSubmissionAccess(context.Context, int32, int32, int32, int32) error {
	return nil
}
func (r *assessmentRepoStub) CourseTrainerAssigned(context.Context, int32, int32, int32) (bool, error) {
	return r.assigned, nil
}
func (*assessmentRepoStub) CreateSubmission(_ context.Context, orgID, courseID, lessonID, userID int32, text string, url *string) (*domainassessment.Submission, error) {
	return &domainassessment.Submission{ID: 1, OrganizationID: orgID, CourseID: courseID, LessonID: lessonID, StudentID: userID, Attempt: 1, ResponseText: text, ResponseURL: url, Status: domainassessment.StatusSubmitted}, nil
}
func (*assessmentRepoStub) ListStudentSubmissions(context.Context, int32, int32, int32, int32) ([]domainassessment.Submission, error) {
	return []domainassessment.Submission{}, nil
}
func (*assessmentRepoStub) ListCourseSubmissions(context.Context, int32, int32, int32, PageQuery) ([]domainassessment.Submission, int64, error) {
	return []domainassessment.Submission{}, 0, nil
}
func (*assessmentRepoStub) ReviewSubmission(_ context.Context, orgID, courseID int32, id int64, reviewerID int32, score float64, feedback string) (*domainassessment.Submission, error) {
	item := &domainassessment.Submission{ID: id, OrganizationID: orgID, CourseID: courseID, Status: domainassessment.StatusSubmitted}
	if err := item.Review(reviewerID, score, feedback, time.Now()); err != nil {
		return nil, err
	}
	return item, nil
}

func TestOnlyStudentsCanSubmit(t *testing.T) {
	repo := &assessmentRepoStub{role: "student"}
	service := NewService(repo, nil)
	item, err := service.Submit(context.Background(), 10, 20, 30, 40, SubmitInput{ResponseText: "answer"})
	if err != nil || item.Attempt != 1 || item.ResponseText != "answer" {
		t.Fatalf("submission = %+v, err=%v", item, err)
	}
	repo.role = "trainer"
	if _, err := service.Submit(context.Background(), 10, 20, 30, 40, SubmitInput{ResponseText: "answer"}); err == nil {
		t.Fatal("trainer must not submit as a student")
	}
}

func TestOnlyAssignedCourseTrainersCanReview(t *testing.T) {
	repo := &assessmentRepoStub{role: "trainer"}
	service := NewService(repo, nil)
	if _, err := service.Review(context.Background(), 10, 20, 30, 40, ReviewInput{Score: 90}); err == nil {
		t.Fatal("unassigned trainer must not review a course submission")
	}
	repo.assigned = true
	item, err := service.Review(context.Background(), 10, 20, 30, 40, ReviewInput{Score: 90, Feedback: "Good"})
	if err != nil || item.Status != domainassessment.StatusReviewed {
		t.Fatalf("reviewed submission = %+v, err=%v", item, err)
	}
}
