CREATE TABLE cohorts (
    id serial PRIMARY KEY,
    organization_id integer NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    course_id integer NOT NULL REFERENCES courses(id) ON DELETE RESTRICT,
    name varchar(160) NOT NULL,
    starts_at timestamptz NULL,
    ends_at timestamptz NULL,
    status varchar(24) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'active', 'completed', 'cancelled')),
    created_by integer NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at >= starts_at),
    UNIQUE (id, organization_id, course_id)
);

CREATE INDEX idx_cohorts_org_status ON cohorts (organization_id, status);
CREATE INDEX idx_cohorts_course_id ON cohorts (course_id);

CREATE TABLE cohort_instructors (
    cohort_id integer NOT NULL REFERENCES cohorts(id) ON DELETE RESTRICT,
    trainer_id integer NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (cohort_id, trainer_id)
);

CREATE INDEX idx_cohort_instructors_trainer_id ON cohort_instructors (trainer_id);

CREATE TABLE cohort_enrollments (
    id bigserial PRIMARY KEY,
    cohort_id integer NOT NULL REFERENCES cohorts(id) ON DELETE RESTRICT,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status varchar(24) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'withdrawn')),
    enrolled_by integer NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    enrolled_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    withdrawn_by integer NULL REFERENCES users(id) ON DELETE RESTRICT,
    withdrawn_at timestamptz NULL,
    withdrawal_reason varchar(500) NULL,
    CHECK ((status = 'active' AND withdrawn_by IS NULL AND withdrawn_at IS NULL)
        OR (status = 'withdrawn' AND withdrawn_by IS NOT NULL AND withdrawn_at IS NOT NULL))
);

CREATE UNIQUE INDEX idx_cohort_enrollments_one_active
    ON cohort_enrollments (cohort_id, user_id) WHERE status = 'active';
CREATE INDEX idx_cohort_enrollments_user_status ON cohort_enrollments (user_id, status);
