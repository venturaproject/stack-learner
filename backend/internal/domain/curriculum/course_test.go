package curriculum

import "testing"

func TestValidateCourse(t *testing.T) {
	if err := ValidateCourse("Go para principiantes", "Curso introductorio"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCourse(" ", ""); err == nil {
		t.Fatal("expected empty title to be rejected")
	}
	if err := ValidateCourse("x", string(make([]rune, 5001))); err == nil {
		t.Fatal("expected oversized description to be rejected")
	}
}

func TestValidateLessonKind(t *testing.T) {
	if err := ValidateLesson("Introducción", "text"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateLesson("Introducción", "executable"); err == nil {
		t.Fatal("expected unsupported lesson kind to be rejected")
	}
}

func TestCoursePublishRequiresDraftAndLesson(t *testing.T) {
	course := Course{Status: CourseDraft}
	if err := course.Publish(0); err == nil {
		t.Fatal("expected an empty course to remain unpublished")
	}
	if course.Status != CourseDraft {
		t.Fatalf("status changed after failed publish: %q", course.Status)
	}
	if err := course.Publish(1); err != nil {
		t.Fatalf("publish course: %v", err)
	}
	if course.Status != CoursePublished {
		t.Fatalf("got status %q, want %q", course.Status, CoursePublished)
	}
	if err := course.Publish(1); err == nil {
		t.Fatal("expected publishing an already-published course to fail")
	}
}

func TestEnsureDraft(t *testing.T) {
	if err := (Course{Status: CourseDraft}).EnsureDraft(); err != nil {
		t.Fatalf("draft course should be mutable: %v", err)
	}
	if err := (Course{Status: CoursePublished}).EnsureDraft(); err == nil {
		t.Fatal("published course should not be mutable")
	}
}
