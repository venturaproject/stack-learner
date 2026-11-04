//go:build integration

package persistence

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	appassessment "boilerplate-learner/internal/application/assessment"
	"boilerplate-learner/internal/application/learning"
	domaincohorts "boilerplate-learner/internal/domain/cohorts"
	"boilerplate-learner/internal/domain/curriculum"
	"boilerplate-learner/internal/infrastructure/persistence/models"
	"boilerplate-learner/internal/shared/apperrors"
)

func TestLearningRepositoryScopesCoursesToOrganization(t *testing.T) {
	db := testDB(t)
	if err := db.AutoMigrate(&models.Organization{}, &models.OrganizationMembership{}, &models.Course{}, &models.CourseModule{}, &models.Lesson{}); err != nil {
		t.Fatalf("migrate learning models: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM assessment_submissions")
		db.Exec("DELETE FROM student_lesson_progress")
		db.Exec("DELETE FROM cohort_enrollments")
		db.Exec("DELETE FROM cohort_instructors")
		db.Exec("DELETE FROM cohorts")
		db.Exec("DELETE FROM lessons")
		db.Exec("DELETE FROM course_modules")
		db.Exec("DELETE FROM courses")
		db.Exec("DELETE FROM organization_memberships")
		db.Exec("DELETE FROM organizations")
	})

	ctx := context.Background()
	repo := NewLearningRepository(db)
	service := learning.NewService(repo, NewTxManager(db))
	ownerA, ownerB := int32(8_100_001), int32(8_100_002)
	orgA, err := service.CreateOrganization(ctx, ownerA, "Tenant A", fmt.Sprintf("tenant-a-%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	orgB, err := service.CreateOrganization(ctx, ownerB, "Tenant B", fmt.Sprintf("tenant-b-%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	courseA, err := repo.CreateCourse(ctx, ownerA, orgA.ID, "Curso privado A", "")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repo.GetCourse(ctx, orgB.ID, courseA.ID); apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("cross-tenant detail should be hidden, got %v", err)
	}
	visible, _, err := repo.ListCourses(ctx, orgB.ID, false, 0, 0, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(visible) != 0 {
		t.Fatalf("tenant B can see tenant A course: %+v", visible)
	}

}

func TestLearningUseCasesAuthorizeAndMutateWithinTransaction(t *testing.T) {
	db := testDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("apply learning migrations: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Organization{}, &models.OrganizationMembership{}, &models.Course{}, &models.CourseModule{}, &models.Lesson{}); err != nil {
		t.Fatalf("migrate learning models: %v", err)
	}
	ownerID := int32(time.Now().UnixNano()%900_000_000) + 1_000_000_000
	studentID, trainerID := ownerID+1, ownerID+2
	t.Cleanup(func() {
		db.Exec("DELETE FROM assessment_submissions")
		db.Exec("DELETE FROM student_lesson_progress")
		db.Exec("DELETE FROM cohort_enrollments")
		db.Exec("DELETE FROM cohort_instructors")
		db.Exec("DELETE FROM cohorts")
		db.Exec("DELETE FROM lessons")
		db.Exec("DELETE FROM course_modules")
		db.Exec("DELETE FROM courses")
		db.Exec("DELETE FROM organization_memberships")
		db.Exec("DELETE FROM course_trainers")
		db.Exec("DELETE FROM organizations")
		db.Exec("DELETE FROM users WHERE id IN (?, ?, ?)", ownerID, studentID, trainerID)
	})

	ctx := context.Background()
	ownerEmail := fmt.Sprintf("owner-%d@example.test", time.Now().UnixNano())
	if err := db.Create(&models.User{ID: ownerID, Name: "Learning Owner", Email: ownerEmail, PasswordHash: "test-hash"}).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewLearningRepository(db)
	service := learning.NewService(repo, NewTxManager(db))
	org, err := service.CreateOrganization(ctx, ownerID, "Learning Tx", fmt.Sprintf("learning-tx-%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	studentEmail := fmt.Sprintf("student-%d@example.test", time.Now().UnixNano())
	if err := db.Create(&models.User{ID: studentID, Name: "Learning Student", Email: studentEmail, PasswordHash: "test-hash"}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddStudent(ctx, ownerID, org.ID, studentEmail); err != nil {
		t.Fatalf("add student: %v", err)
	}
	trainerEmail := fmt.Sprintf("trainer-%d@example.test", time.Now().UnixNano())
	if err := db.Create(&models.User{ID: trainerID, Name: "Learning Trainer", Email: trainerEmail, PasswordHash: "test-hash"}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddMember(ctx, ownerID, org.ID, trainerEmail, "trainer"); err != nil {
		t.Fatalf("add trainer: %v", err)
	}

	course, err := service.CreateCourse(ctx, ownerID, org.ID, "Transactional course", "")
	if err != nil {
		t.Fatal(err)
	}
	course, err = service.UpdateCourse(ctx, ownerID, org.ID, course.ID, "Updated transactional course", "Updated description")
	if err != nil || course.Title != "Updated transactional course" {
		t.Fatalf("update course: course=%+v err=%v", course, err)
	}
	if err := service.AssignCourseTrainer(ctx, ownerID, org.ID, course.ID, trainerID); err != nil {
		t.Fatalf("assign trainer: %v", err)
	}
	assignedTrainers, err := service.ListCourseTrainers(ctx, ownerID, org.ID, course.ID)
	if err != nil || len(assignedTrainers) != 1 || !assignedTrainers[0].Assigned {
		t.Fatalf("list assigned trainers: trainers=%+v err=%v", assignedTrainers, err)
	}
	if _, err := service.UpdateCourse(ctx, trainerID, org.ID, course.ID, "Trainer-edited course", "Updated description"); err != nil {
		t.Fatalf("assigned trainer should edit the course: %v", err)
	}
	trainerCreatedCourse, err := service.CreateCourse(ctx, trainerID, org.ID, "Trainer-owned draft", "")
	if err != nil {
		t.Fatalf("trainer should create course: %v", err)
	}
	if _, err := service.UpdateCourse(ctx, trainerID, org.ID, trainerCreatedCourse.ID, "Trainer-owned draft updated", ""); err != nil {
		t.Fatalf("trainer-created course should be assigned to its creator: %v", err)
	}
	if err := service.DeleteCourse(ctx, trainerID, org.ID, trainerCreatedCourse.ID); err != nil {
		t.Fatalf("delete trainer-created draft: %v", err)
	}
	otherCourse, err := service.CreateCourse(ctx, ownerID, org.ID, "Another course", "")
	if err != nil {
		t.Fatal(err)
	}
	trainerCourses, err := service.ListCourses(ctx, trainerID, org.ID, learning.PageQuery{Page: 1, PerPage: 10})
	if err != nil || trainerCourses.Total != 1 || trainerCourses.Items[0].ID != course.ID {
		t.Fatalf("trainer should only list assigned courses: page=%+v err=%v", trainerCourses, err)
	}
	if _, err := service.GetCourse(ctx, trainerID, org.ID, otherCourse.ID); apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("unassigned course should be hidden from trainer, got %v", err)
	}
	firstPage, err := service.ListCourses(ctx, ownerID, org.ID, learning.PageQuery{Page: 1, PerPage: 1, Search: "course"})
	if err != nil || firstPage.Total != 2 || firstPage.TotalPages != 2 || len(firstPage.Items) != 1 {
		t.Fatalf("course pagination/search failed: page=%+v err=%v", firstPage, err)
	}
	moduleOne, err := service.AddModule(ctx, ownerID, org.ID, course.ID, "Module one")
	if err != nil {
		t.Fatal(err)
	}
	moduleTwo, err := service.AddModule(ctx, ownerID, org.ID, course.ID, "Module two")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateModule(ctx, ownerID, org.ID, course.ID, moduleTwo.ID, "Renamed module"); err != nil {
		t.Fatal(err)
	}
	lessonOne, err := service.AddLesson(ctx, ownerID, org.ID, course.ID, moduleOne.ID, "Lesson one", "text", "")
	if err != nil {
		t.Fatal(err)
	}
	lessonTwo, err := service.AddLesson(ctx, ownerID, org.ID, course.ID, moduleOne.ID, "Lesson two", "text", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateLesson(ctx, ownerID, org.ID, course.ID, moduleOne.ID, lessonTwo.ID, "Updated lesson", "text", "content"); err != nil {
		t.Fatal(err)
	}
	if err := service.ReorderLessons(ctx, ownerID, org.ID, course.ID, moduleOne.ID, []int32{lessonTwo.ID, lessonOne.ID}); err != nil {
		t.Fatal(err)
	}
	modules, err := repo.GetCourse(ctx, org.ID, course.ID)
	if err != nil {
		t.Fatal(err)
	}
	if modules.Modules[0].Lessons[0].ID != lessonTwo.ID || modules.Modules[0].Lessons[0].Position != 1 {
		t.Fatalf("lesson reorder failed: %+v", modules.Modules[0].Lessons)
	}
	if err := service.DeleteLesson(ctx, ownerID, org.ID, course.ID, moduleOne.ID, lessonTwo.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.ReorderModules(ctx, ownerID, org.ID, course.ID, []int32{moduleTwo.ID, moduleOne.ID}); err != nil {
		t.Fatal(err)
	}
	moduleToDelete, err := service.AddModule(ctx, ownerID, org.ID, course.ID, "Temporary module")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	addErrors := make(chan error, 6)
	for i := 0; i < cap(addErrors); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := service.AddLesson(ctx, ownerID, org.ID, course.ID, moduleToDelete.ID, fmt.Sprintf("Concurrent lesson %d", index), "text", "")
			addErrors <- err
		}(i)
	}
	wg.Wait()
	close(addErrors)
	for err := range addErrors {
		if err != nil {
			t.Fatalf("concurrent lesson insert: %v", err)
		}
	}
	concurrentCourse, err := repo.GetCourse(ctx, org.ID, course.ID)
	if err != nil {
		t.Fatal(err)
	}
	var concurrentModule *curriculum.Module
	for i := range concurrentCourse.Modules {
		if concurrentCourse.Modules[i].ID == moduleToDelete.ID {
			concurrentModule = &concurrentCourse.Modules[i]
		}
	}
	if concurrentModule == nil || len(concurrentModule.Lessons) != cap(addErrors) {
		t.Fatalf("concurrent lessons missing: %+v", concurrentCourse.Modules)
	}
	positions := make(map[int]bool)
	for _, lesson := range concurrentModule.Lessons {
		if positions[lesson.Position] {
			t.Fatalf("duplicate lesson position %d", lesson.Position)
		}
		positions[lesson.Position] = true
	}
	for position := 1; position <= len(concurrentModule.Lessons); position++ {
		if !positions[position] {
			t.Fatalf("missing lesson position %d: %+v", position, positions)
		}
	}
	if err := service.DeleteModule(ctx, ownerID, org.ID, course.ID, moduleToDelete.ID); err != nil {
		t.Fatal(err)
	}
	sequenceLesson, err := service.AddLesson(ctx, ownerID, org.ID, course.ID, moduleOne.ID, "Task lesson", "task", "Describe your solution")
	if err != nil {
		t.Fatalf("add second sequential lesson: %v", err)
	}
	modules, err = repo.GetCourse(ctx, org.ID, course.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(modules.Modules) != 2 || modules.Modules[0].ID != moduleTwo.ID || modules.Modules[0].Position != 1 || modules.Modules[1].ID != moduleOne.ID || modules.Modules[1].Position != 2 {
		t.Fatalf("module reorder/delete failed: %+v", modules.Modules)
	}
	if _, err := service.PublishCourse(ctx, ownerID, org.ID, course.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetCourseProgress(ctx, studentID, org.ID, course.ID); apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("student without an active cohort enrollment must not access progress, got %v", err)
	}
	if _, err := service.CompleteLesson(ctx, studentID, org.ID, course.ID, lessonOne.ID); apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("student without an active cohort enrollment must not complete lessons, got %v", err)
	}
	cohort, err := service.CreateCohort(ctx, ownerID, org.ID, course.ID, learning.CreateCohortInput{Name: "Spring cohort"})
	if err != nil || cohort.Status != domaincohorts.StatusDraft {
		t.Fatalf("create draft cohort: cohort=%+v err=%v", cohort, err)
	}
	if _, err := service.EnrollStudent(ctx, ownerID, org.ID, cohort.ID, studentID); apperrors.KindOf(err) != apperrors.KindConflict {
		t.Fatalf("enrollment into draft cohort should fail, got %v", err)
	}
	if err := service.TransitionCohort(ctx, ownerID, org.ID, cohort.ID, domaincohorts.StatusActive); err != nil {
		t.Fatalf("activate cohort: %v", err)
	}
	if _, err := service.ListCohortEnrollments(ctx, trainerID, org.ID, cohort.ID, learning.PageQuery{Page: 1, PerPage: 25}); apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("unassigned trainer should not list enrollments, got %v", err)
	}
	if err := service.AssignCohortInstructor(ctx, ownerID, org.ID, cohort.ID, trainerID); err != nil {
		t.Fatalf("assign cohort instructor: %v", err)
	}
	enrollment, err := service.EnrollStudent(ctx, ownerID, org.ID, cohort.ID, studentID)
	if err != nil || enrollment.Status != domaincohorts.EnrollmentActive {
		t.Fatalf("enroll student: enrollment=%+v err=%v", enrollment, err)
	}
	if _, err := service.EnrollStudent(ctx, ownerID, org.ID, cohort.ID, studentID); apperrors.KindOf(err) != apperrors.KindConflict {
		t.Fatalf("duplicate active enrollment should conflict, got %v", err)
	}
	if err := service.WithdrawStudent(ctx, trainerID, org.ID, cohort.ID, studentID, "schedule conflict"); err != nil {
		t.Fatalf("assigned trainer withdrawal: %v", err)
	}
	if _, err := service.GetCourse(ctx, studentID, org.ID, course.ID); apperrors.KindOf(err) != apperrors.KindNotFound {
		t.Fatalf("withdrawn student should not access course, got %v", err)
	}
	if _, err := service.EnrollStudent(ctx, ownerID, org.ID, cohort.ID, studentID); err != nil {
		t.Fatalf("re-enroll after withdrawal should preserve history: %v", err)
	}
	studentCohorts, err := service.ListCohorts(ctx, studentID, org.ID, course.ID, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil || studentCohorts.Total != 1 || studentCohorts.Items[0].ID != cohort.ID {
		t.Fatalf("student should only see their enrolled cohort: cohorts=%+v err=%v", studentCohorts, err)
	}
	if _, err := service.GetCourseProgress(ctx, ownerID, org.ID, course.ID); apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("organization admin must not read student progress endpoint, got %v", err)
	}
	progress, err := service.GetCourseProgress(ctx, studentID, org.ID, course.ID)
	if err != nil || progress.TotalLessons != 2 || progress.CompletedLessons != 0 || progress.Percent != 0 || !progress.Lessons[0].Unlocked || progress.Lessons[1].Unlocked {
		t.Fatalf("initial course progress = %+v, err=%v", progress, err)
	}
	if _, err := service.CompleteLesson(ctx, studentID, org.ID, course.ID, sequenceLesson.ID); apperrors.KindOf(err) != apperrors.KindConflict {
		t.Fatalf("later lesson must stay locked until the previous lesson is complete, got %v", err)
	}
	studentCourse, err := service.GetCourse(ctx, studentID, org.ID, course.ID)
	if err != nil {
		t.Fatal(err)
	}
	if studentCourse.Modules[1].Lessons[1].Content != "" {
		t.Fatalf("locked lesson content must not be exposed to the student: %+v", studentCourse.Modules[1].Lessons[1])
	}
	completed, err := service.CompleteLesson(ctx, studentID, org.ID, course.ID, lessonOne.ID)
	if err != nil || completed.CompletedAt == nil {
		t.Fatalf("complete lesson: progress=%+v err=%v", completed, err)
	}
	completedAgain, err := service.CompleteLesson(ctx, studentID, org.ID, course.ID, lessonOne.ID)
	if err != nil || completedAgain.CompletedAt == nil || !completedAgain.CompletedAt.Equal(*completed.CompletedAt) {
		t.Fatalf("completing same lesson should be idempotent: first=%+v second=%+v err=%v", completed, completedAgain, err)
	}
	progress, err = service.GetCourseProgress(ctx, studentID, org.ID, course.ID)
	if err != nil || progress.CompletedLessons != 1 || progress.Percent != 50 || !progress.Lessons[1].Unlocked {
		t.Fatalf("updated course progress = %+v, err=%v", progress, err)
	}
	if _, err := service.CompleteLesson(ctx, studentID, org.ID, course.ID, sequenceLesson.ID); apperrors.KindOf(err) != apperrors.KindConflict {
		t.Fatalf("task lesson should require a submission before completion, got %v", err)
	}
	assessmentService := appassessment.NewService(NewAssessmentRepository(db), NewTxManager(db))
	firstSubmission, err := assessmentService.Submit(ctx, studentID, org.ID, course.ID, sequenceLesson.ID, appassessment.SubmitInput{ResponseText: "first answer"})
	if err != nil || firstSubmission.Attempt != 1 || firstSubmission.Status != "submitted" {
		t.Fatalf("first task submission = %+v, err=%v", firstSubmission, err)
	}
	secondSubmission, err := assessmentService.Submit(ctx, studentID, org.ID, course.ID, sequenceLesson.ID, appassessment.SubmitInput{ResponseURL: "https://example.test/work"})
	if err != nil || secondSubmission.Attempt != 2 {
		t.Fatalf("second task submission = %+v, err=%v", secondSubmission, err)
	}
	if _, err := service.CompleteLesson(ctx, studentID, org.ID, course.ID, sequenceLesson.ID); err != nil {
		t.Fatalf("task lesson should unlock completion after a submission: %v", err)
	}
	if _, err := assessmentService.Review(ctx, studentID, org.ID, course.ID, firstSubmission.ID, appassessment.ReviewInput{Score: 90, Feedback: "Well done"}); apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("student should not review submissions, got %v", err)
	}
	reviewed, err := assessmentService.Review(ctx, trainerID, org.ID, course.ID, firstSubmission.ID, appassessment.ReviewInput{Score: 92.5, Feedback: "Well done"})
	if err != nil || reviewed.Status != "reviewed" || reviewed.Score == nil || *reviewed.Score != 92.5 || reviewed.ReviewedAt == nil {
		t.Fatalf("reviewed submission = %+v, err=%v", reviewed, err)
	}
	if _, err := assessmentService.Review(ctx, trainerID, org.ID, course.ID, firstSubmission.ID, appassessment.ReviewInput{Score: 95}); apperrors.KindOf(err) != apperrors.KindConflict {
		t.Fatalf("review should not overwrite a prior result, got %v", err)
	}
	thirdSubmission, err := assessmentService.Submit(ctx, studentID, org.ID, course.ID, sequenceLesson.ID, appassessment.SubmitInput{ResponseText: "revised answer"})
	if err != nil || thirdSubmission.Attempt != 3 {
		t.Fatalf("third task submission should preserve attempt history: %+v, err=%v", thirdSubmission, err)
	}
	studentAttempts, err := assessmentService.ListStudentSubmissions(ctx, studentID, org.ID, course.ID, sequenceLesson.ID)
	if err != nil || len(studentAttempts) != 3 || studentAttempts[0].Attempt != 3 || studentAttempts[2].Status != "reviewed" {
		t.Fatalf("student submission history = %+v, err=%v", studentAttempts, err)
	}
	reviewQueue, err := assessmentService.ListCourseSubmissions(ctx, trainerID, org.ID, course.ID, appassessment.PageQuery{Page: 1, PerPage: 25})
	if err != nil || reviewQueue.Total != 3 {
		t.Fatalf("teacher review queue = %+v, err=%v", reviewQueue, err)
	}
	progress, err = service.GetCourseProgress(ctx, studentID, org.ID, course.ID)
	if err != nil || progress.CompletedLessons != 2 || progress.Percent != 100 {
		t.Fatalf("completed course progress = %+v, err=%v", progress, err)
	}
	if _, err := service.CompleteLesson(ctx, trainerID, org.ID, course.ID, lessonOne.ID); apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("trainer must not complete a lesson for a student, got %v", err)
	}
	enrollments, err := service.ListCohortEnrollments(ctx, trainerID, org.ID, cohort.ID, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil || enrollments.Total != 2 {
		t.Fatalf("enrollment history = %+v, err=%v; want active and withdrawn records", enrollments, err)
	}
	if enrollments.Items[0].Status != domaincohorts.EnrollmentActive || enrollments.Items[1].Status != domaincohorts.EnrollmentWithdrawn || enrollments.Items[1].WithdrawalReason == nil || *enrollments.Items[1].WithdrawalReason != "schedule conflict" {
		t.Fatalf("enrollment history details = %+v", enrollments.Items)
	}
	adminCourses, err := service.ListCourses(ctx, ownerID, org.ID, learning.PageQuery{Page: 1, PerPage: 10})
	if err != nil || adminCourses.Summary == nil || adminCourses.Summary.CourseCount != 2 || adminCourses.Summary.PublishedCount != 1 || adminCourses.Summary.ModuleCount != 2 {
		t.Fatalf("course summary mismatch: summary=%+v err=%v", adminCourses.Summary, err)
	}

	studentCourses, err := service.ListCourses(ctx, studentID, org.ID, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(studentCourses.Items) != 1 || studentCourses.Items[0].Status != curriculum.CoursePublished {
		t.Fatalf("student should only receive published courses, got %+v", studentCourses)
	}
	if _, err := service.GetCourse(ctx, studentID, org.ID, course.ID); err != nil {
		t.Fatalf("student with an active cohort enrollment should access the course: %v", err)
	}
	studentList, err := service.ListStudents(ctx, ownerID, org.ID, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil || len(studentList.Items) != 1 || studentList.Items[0].Email != studentEmail {
		t.Fatalf("list students: students=%+v err=%v", studentList, err)
	}
	if err := service.RemoveStudent(ctx, ownerID, org.ID, studentID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetCourseProgress(ctx, studentID, org.ID, course.ID); apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("removed member should not read course progress, got %v", err)
	}
	enrollments, err = service.ListCohortEnrollments(ctx, ownerID, org.ID, cohort.ID, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil || enrollments.Total != 2 || enrollments.Items[0].Status != domaincohorts.EnrollmentWithdrawn {
		t.Fatalf("removing an organization member should preserve cohort history and withdraw active enrollment: %+v err=%v", enrollments, err)
	}
	if _, err := service.GetCourse(ctx, studentID, org.ID, course.ID); apperrors.KindOf(err) != apperrors.KindForbidden {
		t.Fatalf("removed organization member should lose course access, got %v", err)
	}
	studentList, err = service.ListStudents(ctx, ownerID, org.ID, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil || len(studentList.Items) != 0 {
		t.Fatalf("remove student: students=%+v err=%v", studentList, err)
	}
	if _, err := service.UpdateCourse(ctx, ownerID, org.ID, course.ID, "Cannot update published", ""); err == nil {
		t.Fatal("published course must reject updates")
	}
	if err := service.DeleteCourse(ctx, ownerID, org.ID, course.ID); err == nil {
		t.Fatal("published course must reject deletion")
	}
	updatedOrg, err := service.UpdateOrganization(ctx, ownerID, org.ID, "Renamed organization", "renamed-learning-tx")
	if err != nil || updatedOrg.Name != "Renamed organization" {
		t.Fatalf("update organization: org=%+v err=%v", updatedOrg, err)
	}
	if err := service.ArchiveOrganization(ctx, ownerID, org.ID); err != nil {
		t.Fatal(err)
	}
	orgs, err := service.ListOrganizations(ctx, ownerID, learning.PageQuery{Page: 1, PerPage: 25})
	if err != nil || len(orgs.Items) != 1 || orgs.Items[0].ArchivedAt == nil {
		t.Fatalf("archive organization: organizations=%+v err=%v", orgs, err)
	}
	if err := service.RestoreOrganization(ctx, ownerID, org.ID); err != nil {
		t.Fatalf("restore organization: %v", err)
	}
}
