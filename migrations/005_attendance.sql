-- Attendance records keyed by class session (subject + start time), not date-only.
-- Apply manually, e.g.:
--   psql "$DATABASE_URL" -f migrations/005_attendance.sql
--
-- If you already applied an older date-only version of this file, also run:
--   psql "$DATABASE_URL" -f migrations/006_attendance_class_session.sql

CREATE TABLE IF NOT EXISTS attendance_records (
    id                 TEXT PRIMARY KEY,
    user_id            TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id         TEXT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    session_starts_at  TIMESTAMPTZ NOT NULL,
    session_ends_at    TIMESTAMPTZ,
    status             TEXT NOT NULL DEFAULT 'present',
    note               TEXT NOT NULL DEFAULT '',
    marked_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT attendance_status_check
        CHECK (status IN ('present', 'absent', 'late')),
    CONSTRAINT attendance_user_subject_session_unique
        UNIQUE (user_id, subject_id, session_starts_at)
);

CREATE INDEX IF NOT EXISTS idx_attendance_user_id
    ON attendance_records(user_id);

CREATE INDEX IF NOT EXISTS idx_attendance_subject_id
    ON attendance_records(subject_id);

CREATE INDEX IF NOT EXISTS idx_attendance_user_session_starts
    ON attendance_records(user_id, session_starts_at DESC);

CREATE INDEX IF NOT EXISTS idx_attendance_user_subject
    ON attendance_records(user_id, subject_id);
