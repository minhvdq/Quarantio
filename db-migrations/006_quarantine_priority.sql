ALTER TABLE quarantine
    ADD COLUMN IF NOT EXISTS priority VARCHAR(10) NOT NULL DEFAULT 'medium'
    CHECK (priority IN ('medium', 'high'));
