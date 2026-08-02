-- User-defined exam types (extends the built-in defaults).
-- Apply manually, e.g.:
--   psql "$DATABASE_URL" -f migrations/003_exam_types.sql

CREATE TABLE IF NOT EXISTS exam_types (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exam_types_user_name_lower
    ON exam_types (user_id, lower(name));

CREATE INDEX IF NOT EXISTS idx_exam_types_user_id ON exam_types(user_id);
