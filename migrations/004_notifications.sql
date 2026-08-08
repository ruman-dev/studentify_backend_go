-- Notifications for user alerts.
-- Apply manually, e.g.:
--   psql "$DATABASE_URL" -f migrations/004_notifications.sql

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
