CREATE TABLE assessment_submissions (
    id bigserial PRIMARY KEY,
    organization_id integer NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    course_id integer NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    lesson_id integer NOT NULL REFERENCES lessons(id) ON DELETE RESTRICT,
    student_id integer NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    attempt_number integer NOT NULL CHECK (attempt_number > 0),
    response_text text NOT NULL DEFAULT '',
    response_url varchar(2048) NULL,
    status varchar(24) NOT NULL DEFAULT 'submitted'
        CHECK (status IN ('submitted', 'reviewed')),
    score numeric(5,2) NULL CHECK (score IS NULL OR (score >= 0 AND score <= 100)),
    feedback text NULL,
    submitted_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_by integer NULL REFERENCES users(id) ON DELETE RESTRICT,
    reviewed_at timestamptz NULL,
    CHECK (length(btrim(response_text)) > 0 OR response_url IS NOT NULL),
    CHECK (
        (status = 'submitted' AND score IS NULL AND reviewed_by IS NULL AND reviewed_at IS NULL)
        OR (status = 'reviewed' AND score IS NOT NULL AND reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)
    ),
    UNIQUE (student_id, lesson_id, attempt_number)
);

CREATE INDEX idx_assessment_submissions_course_status
    ON assessment_submissions (organization_id, course_id, status, submitted_at DESC);
CREATE INDEX idx_assessment_submissions_lesson
    ON assessment_submissions (lesson_id, student_id, attempt_number DESC);
CREATE INDEX idx_assessment_submissions_student
    ON assessment_submissions (student_id, submitted_at DESC);
