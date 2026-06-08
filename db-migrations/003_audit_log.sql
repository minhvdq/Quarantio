CREATE TABLE IF NOT EXISTS audit_log (
    id               UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id       UUID    NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    tenant_id        UUID    NOT NULL REFERENCES tenants(id),
    email_from       VARCHAR(255),
    email_to         TEXT[],
    email_subject    VARCHAR(500),
    verdict          VARCHAR(10) NOT NULL CHECK (verdict IN ('CLEAN', 'LOW', 'MEDIUM', 'HIGH')),
    violations       JSONB,
    gemini_reasoning TEXT,
    action_taken     VARCHAR(50),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_tenant_time
    ON audit_log (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_request_id
    ON audit_log (request_id);
