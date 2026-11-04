CREATE TABLE student_lesson_progress (
    user_id integer NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    lesson_id integer NOT NULL REFERENCES lessons(id) ON DELETE RESTRICT,
    completed_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, lesson_id)
);

CREATE INDEX idx_student_lesson_progress_lesson_id
    ON student_lesson_progress (lesson_id);
