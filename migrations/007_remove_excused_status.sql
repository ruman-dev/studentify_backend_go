-- Remove unused "excused" attendance status.
-- Existing excused rows are converted to absent.
-- Apply: psql "$DATABASE_URL" -f migrations/007_remove_excused_status.sql

UPDATE attendance_records
SET status = 'absent',
    updated_at = NOW()
WHERE status = 'excused';

ALTER TABLE attendance_records
    DROP CONSTRAINT IF EXISTS attendance_status_check;

ALTER TABLE attendance_records
    ADD CONSTRAINT attendance_status_check
        CHECK (status IN ('present', 'absent', 'late'));
