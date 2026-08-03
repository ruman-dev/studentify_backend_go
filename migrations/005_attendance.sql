-- Attendance records for self-tracked class sessions.
-- Apply manually, e.g.:
--   psql "$DATABASE_URL" -f migrations/005_attendance.sql

CREATE TABLE IF NOT EXISTS attendance_records (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id    TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    session_date  DATE NOT NULL,
    status        TEXT NOT NULL DEFAULT 'present',
    note          TEXT NOT NULL DEFAULT '',
    marked_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT attendance_status_check
        CHECK (status IN ('present', 'absent', 'late', 'excused')),
    CONSTRAINT attendance_user_subject_date_unique
        UNIQUE (user_id, subject_id, session_date)
);

CREATE INDEX IF NOT EXISTS idx_attendance_user_id
    ON attendance_records(user_id);

CREATE INDEX IF NOT EXISTS idx_attendance_subject_id
    ON attendance_records(subject_id);

CREATE INDEX IF NOT EXISTS idx_attendance_user_session_date
    ON attendance_records(user_id, session_date DESC);

CREATE INDEX IF NOT EXISTS idx_attendance_user_subject
    ON attendance_records(user_id, subject_id);
