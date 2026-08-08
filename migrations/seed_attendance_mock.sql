-- Seed mock attendance for development / QA.
-- Targets the verified demo user (ruman.cse49@gmail.com) and subjects
-- that already have weekly schedules.
--
--   psql "$DATABASE_URL" -f migrations/seed_attendance_mock.sql

DO $$
DECLARE
    v_user_id    TEXT := '56aec7fa-8c6e-4b8f-add4-d09dd56307ee';
    v_ds_id      TEXT := '2741e923-ac7a-436d-9693-bbd169902d9d'; -- Data Structures Mon/Wed 09:00-10:30
    v_it_id      TEXT := '1bfc512e-742f-490f-b703-55e48b3bfee6'; -- IT communication Mon/Tue/Wed 10:00-11:00
    v_tz         TEXT := 'Asia/Dhaka';
    v_day        DATE;
    v_start      TIMESTAMPTZ;
    v_end        TIMESTAMPTZ;
    v_status     TEXT;
    v_weekday    INT;
    v_i          INT;
BEGIN
    -- Clear previous mock rows for this user (idempotent re-seed).
    DELETE FROM attendance_records WHERE user_id = v_user_id;

    -- Walk the last ~5 weeks (including today).
    FOR v_i IN 0..34 LOOP
        v_day := (CURRENT_TIMESTAMP AT TIME ZONE v_tz)::date - v_i;
        v_weekday := EXTRACT(ISODOW FROM v_day)::INT; -- 1=Mon … 7=Sun

        -- Data Structures: Monday / Wednesday 09:00–10:30 Asia/Dhaka
        IF v_weekday IN (1, 3) THEN
            v_start := (v_day + TIME '09:00') AT TIME ZONE v_tz;
            v_end   := (v_day + TIME '10:30') AT TIME ZONE v_tz;

            -- Skip future sessions that have not ended yet.
            IF v_end <= NOW() THEN
                v_status := CASE (EXTRACT(DAY FROM v_day)::INT % 4)
                    WHEN 0 THEN 'absent'
                    WHEN 1 THEN 'late'
                    ELSE 'present'
                END;

                INSERT INTO attendance_records (
                    id, user_id, subject_id, session_starts_at, session_ends_at,
                    status, note, marked_at, created_at, updated_at
                ) VALUES (
                    gen_random_uuid()::text, v_user_id, v_ds_id, v_start, v_end,
                    v_status,
                    CASE v_status
                        WHEN 'absent' THEN 'Missed due to traffic'
                        WHEN 'late' THEN 'Arrived 10 minutes late'
                        ELSE ''
                    END,
                    v_end + INTERVAL '15 minutes',
                    v_end + INTERVAL '15 minutes',
                    v_end + INTERVAL '15 minutes'
                )
                ON CONFLICT (user_id, subject_id, session_starts_at) DO NOTHING;
            END IF;
        END IF;

        -- IT communication: Monday / Tuesday / Wednesday 10:00–11:00 Asia/Dhaka
        IF v_weekday IN (1, 2, 3) THEN
            v_start := (v_day + TIME '10:00') AT TIME ZONE v_tz;
            v_end   := (v_day + TIME '11:00') AT TIME ZONE v_tz;

            IF v_end <= NOW() THEN
                v_status := CASE (EXTRACT(DAY FROM v_day)::INT % 4)
                    WHEN 0 THEN 'late'
                    WHEN 1 THEN 'absent'
                    ELSE 'present'
                END;

                INSERT INTO attendance_records (
                    id, user_id, subject_id, session_starts_at, session_ends_at,
                    status, note, marked_at, created_at, updated_at
                ) VALUES (
                    gen_random_uuid()::text, v_user_id, v_it_id, v_start, v_end,
                    v_status,
                    CASE v_status
                        WHEN 'absent' THEN 'Family emergency'
                        WHEN 'late' THEN 'Lab ran over'
                        ELSE ''
                    END,
                    v_end + INTERVAL '10 minutes',
                    v_end + INTERVAL '10 minutes',
                    v_end + INTERVAL '10 minutes'
                )
                ON CONFLICT (user_id, subject_id, session_starts_at) DO NOTHING;
            END IF;
        END IF;
    END LOOP;
END $$;

-- Quick summary for the seeded user.
SELECT
    s.name,
    s.code,
    COUNT(*) AS sessions,
    COUNT(*) FILTER (WHERE a.status = 'present') AS present,
    COUNT(*) FILTER (WHERE a.status = 'absent') AS absent,
    COUNT(*) FILTER (WHERE a.status = 'late') AS late
FROM attendance_records a
JOIN subjects s ON s.id = a.subject_id
WHERE a.user_id = '56aec7fa-8c6e-4b8f-add4-d09dd56307ee'
GROUP BY s.name, s.code
ORDER BY s.name;
