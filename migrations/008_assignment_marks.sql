-- Add total/obtained marks to assignments (mirrors exams).
-- Apply: psql "$DATABASE_URL" -f migrations/008_assignment_marks.sql

ALTER TABLE assignments
    ADD COLUMN IF NOT EXISTS total_marks DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS obtained_marks DOUBLE PRECISION;
