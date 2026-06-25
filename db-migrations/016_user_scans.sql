-- Per-user scan tracking (display only; quota enforcement stays at tenant level)
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS scans_this_period INTEGER     NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS period_reset_at   TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '1 month');
