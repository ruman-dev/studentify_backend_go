-- Academic modules schema for Studentify (personal student app).
-- Apply manually against your Postgres database, e.g.:
--   psql "$DATABASE_URL" -f migrations/001_academic.sql

CREATE TABLE IF NOT EXISTS student_profiles (
    user_id           TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    date_of_birth     DATE,
    institute_name    TEXT NOT NULL DEFAULT '',
    degree_or_class   TEXT NOT NULL DEFAULT '',
    section           TEXT NOT NULL DEFAULT '',
    student_id_number TEXT NOT NULL DEFAULT '',
    address           TEXT NOT NULL DEFAULT '',
    city              TEXT NOT NULL DEFAULT '',
    country           TEXT NOT NULL DEFAULT '',
    bio               TEXT NOT NULL DEFAULT '',
    others            TEXT NOT NULL DEFAULT '',
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teachers (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    full_name    TEXT NOT NULL,
    email        TEXT NOT NULL DEFAULT '',
    phone        TEXT NOT NULL DEFAULT '',
    department   TEXT NOT NULL DEFAULT '',
    designation  TEXT NOT NULL DEFAULT '',
    notes        TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_teachers_user_id ON teachers(user_id);

CREATE TABLE IF NOT EXISTS subjects (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    teacher_id     TEXT REFERENCES teachers(id) ON DELETE SET NULL,
    name           TEXT NOT NULL,
    code           TEXT NOT NULL DEFAULT '',
    description    TEXT NOT NULL DEFAULT '',
    credit_hours   DOUBLE PRECISION,
    schedule_days  TEXT[] NOT NULL DEFAULT '{}',
    start_time     TEXT NOT NULL DEFAULT '',
    end_time       TEXT NOT NULL DEFAULT '',
    room           TEXT NOT NULL DEFAULT '',
    meeting_link   TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subjects_user_id ON subjects(user_id);
CREATE INDEX IF NOT EXISTS idx_subjects_teacher_id ON subjects(teacher_id);

CREATE TABLE IF NOT EXISTS assignments (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id  TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    due_date    TIMESTAMPTZ NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assignments_user_id ON assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_assignments_subject_id ON assignments(subject_id);

CREATE TABLE IF NOT EXISTS exams (
    id               TEXT PRIMARY KEY,
    user_id          TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id       TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    title            TEXT NOT NULL DEFAULT '',
    exam_type        TEXT NOT NULL DEFAULT '',
    exam_date        TIMESTAMPTZ NOT NULL,
    duration_minutes INTEGER,
    venue            TEXT NOT NULL DEFAULT '',
    total_marks      DOUBLE PRECISION,
    obtained_marks   DOUBLE PRECISION,
    notes            TEXT NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exams_user_id ON exams(user_id);
CREATE INDEX IF NOT EXISTS idx_exams_subject_id ON exams(subject_id);

CREATE TABLE IF NOT EXISTS exam_types (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_types_user_name_lower
    ON exam_types (user_id, lower(name));

CREATE INDEX IF NOT EXISTS idx_exam_types_user_id ON exam_types(user_id);

CREATE TABLE IF NOT EXISTS events (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id  TEXT REFERENCES subjects(id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    location    TEXT NOT NULL DEFAULT '',
    starts_at   TIMESTAMPTZ NOT NULL,
    ends_at     TIMESTAMPTZ,
    event_type  TEXT NOT NULL DEFAULT 'other',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_subject_id ON events(subject_id);

CREATE TABLE IF NOT EXISTS notifications (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title          TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    type           TEXT NOT NULL DEFAULT 'info',
    is_read        BOOLEAN NOT NULL DEFAULT FALSE,
    read_at        TIMESTAMPTZ,
    reference_type TEXT NOT NULL DEFAULT '',
    reference_id   TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_is_read ON notifications(user_id, is_read);
