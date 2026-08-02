-- Add class schedule fields to subjects and optional exam_type on exams.
-- Apply manually if 001_academic.sql was already applied, e.g.:
--   psql "$DATABASE_URL" -f migrations/002_subject_schedule_exam_type.sql

ALTER TABLE subjects
    ADD COLUMN IF NOT EXISTS schedule_days TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS start_time TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS end_time TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS room TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS meeting_link TEXT NOT NULL DEFAULT '';

ALTER TABLE exams
    ADD COLUMN IF NOT EXISTS exam_type TEXT NOT NULL DEFAULT '';

ALTER TABLE exams
    ALTER COLUMN title SET DEFAULT '';
