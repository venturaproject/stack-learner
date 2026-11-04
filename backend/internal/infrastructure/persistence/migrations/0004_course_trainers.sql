CREATE TABLE IF NOT EXISTS course_trainers (
    course_id integer NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    trainer_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (course_id, trainer_id)
);

CREATE INDEX IF NOT EXISTS idx_course_trainers_trainer_id ON course_trainers (trainer_id);
