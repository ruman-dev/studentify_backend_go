-- Migrate attendance from date-only uniqueness to class-session uniqueness.
-- Safe to run if 005 was already applied with the old schema.
--   psql "$DATABASE_URL" -f migrations/006_attendance_class_session.sql

DO $$
BEGIN
    -- Old schema used session_date; new schema uses session_starts_at.
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'attendance_records'
          AND column_name = 'session_date'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'attendance_records'
          AND column_name = 'session_starts_at'
    ) THEN
        ALTER TABLE attendance_records
            ADD COLUMN session_starts_at TIMESTAMPTZ,
            ADD COLUMN session_ends_at TIMESTAMPTZ;

        -- Best-effort backfill: treat old date as midnight UTC (no class time was stored).
        UPDATE attendance_records
        SET session_starts_at = session_date::timestamptz
        WHERE session_starts_at IS NULL;

        ALTER TABLE attendance_records
            ALTER COLUMN session_starts_at SET NOT NULL;

        ALTER TABLE attendance_records
            DROP CONSTRAINT IF EXISTS attendance_user_subject_date_unique;

        ALTER TABLE attendance_records
            DROP COLUMN session_date;

        ALTER TABLE attendance_records
            ADD CONSTRAINT attendance_user_subject_session_unique
            UNIQUE (user_id, subject_id, session_starts_at);

        DROP INDEX IF EXISTS idx_attendance_user_session_date;

        CREATE INDEX IF NOT EXISTS idx_attendance_user_session_starts
            ON attendance_records(user_id, session_starts_at DESC);
    END IF;
END $$;
