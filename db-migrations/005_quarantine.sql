CREATE TABLE IF NOT EXISTS quarantine (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        REFERENCES tenants(id) ON DELETE CASCADE,
    email_from  VARCHAR(255) NOT NULL,
    email_to    VARCHAR(255) NOT NULL,
    subject     VARCHAR(500),
    body        TEXT,
    violations  JSONB,
    reasoning   TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'released', 'rejected')),
    reviewed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quarantine_tenant_status
    ON quarantine (tenant_id, status, created_at DESC);
