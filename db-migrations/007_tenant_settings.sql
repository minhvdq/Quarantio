CREATE TABLE IF NOT EXISTS tenant_settings (
    tenant_id        UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    auto_deliver_low BOOLEAN NOT NULL DEFAULT true,
    retention_days   INTEGER NOT NULL DEFAULT 90,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
