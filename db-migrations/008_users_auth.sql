-- Extend tenants with domain and plan fields
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS domain             VARCHAR(255) UNIQUE,
    ADD COLUMN IF NOT EXISTS domain_verified    BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS plan               VARCHAR(20)  NOT NULL DEFAULT 'free',
    ADD COLUMN IF NOT EXISTS mailbox_count      INTEGER      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS scans_this_period  INTEGER      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS period_reset_at    TIMESTAMPTZ  NOT NULL DEFAULT (NOW() + INTERVAL '1 month'),
    ADD COLUMN IF NOT EXISTS payg_credits       INTEGER      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255),
    ADD COLUMN IF NOT EXISTS stripe_sub_id      VARCHAR(255);

-- Dashboard users (separate from API key tenants)
CREATE TABLE IF NOT EXISTS users (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100) NOT NULL DEFAULT '',
    last_name       VARCHAR(100) NOT NULL DEFAULT '',
    email_verified  BOOLEAN     NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Org membership (user ↔ tenant with role)
CREATE TABLE IF NOT EXISTS org_members (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role            VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('owner','manager','monitor','user')),
    invited_by      UUID        REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_org_members_tenant ON org_members (tenant_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user   ON org_members (user_id);

-- Refresh token sessions
CREATE TABLE IF NOT EXISTS user_sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      CHAR(64)    NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_token ON user_sessions (token_hash);

-- Release requests (user appeals on quarantined emails)
CREATE TABLE IF NOT EXISTS release_requests (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    quarantine_id       UUID        NOT NULL REFERENCES quarantine(id) ON DELETE CASCADE,
    tenant_id           UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    requested_by        UUID        NOT NULL REFERENCES users(id),
    note                TEXT,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','denied')),
    reviewed_by         UUID        REFERENCES users(id),
    reviewed_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_release_requests_tenant ON release_requests (tenant_id);
CREATE INDEX IF NOT EXISTS idx_release_requests_qid    ON release_requests (quarantine_id);
