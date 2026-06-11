-- Allow audit_log and email_history_embeddings to exist without a tenant
-- (emails routed without a tenant_id should still be logged)
ALTER TABLE audit_log ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE email_history_embeddings ALTER COLUMN tenant_id DROP NOT NULL;
