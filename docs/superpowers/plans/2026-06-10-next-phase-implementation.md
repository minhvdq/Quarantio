# GoMailGuard — Next Phase Implementation Plan
**Date:** 2026-06-10  
**Status:** Pending  
**Context:** Post-session plan — captures work discovered through end-to-end testing and product strategy discussion.

---

## Current State (as of 2026-06-10)

Working end-to-end pipeline:
- RabbitMQ → Mistral agent loop → verdict → quarantine table / mail delivery
- Tenant registration, policy upload (Mistral embeddings), audit log API, quarantine review API
- Front-end compliance dashboard (Send Email / Audit Log / Quarantine tabs)
- Quarantine release wired to actual mail delivery via mail-service

Known issues / tech debt:
- `ai-compliance-service/complianceApp` binary is uncommitted (modified but not committed)
- `tenant-service/tenantApp` binary is uncommitted (modified but not committed)
- No sync check endpoint (`POST /v1/check`) — core B2B API product doesn't exist yet
- HIGH verdict goes to quarantine but not flagged as priority — same UX as MEDIUM
- No tenant compliance config (`POST /v1/settings`)
- No data privacy endpoints (retention, erasure, export)
- Quarantine body stored in plaintext — no encryption at rest

---

## Phase 4: Synchronous Check API (P0)

**Why first:** This is the actual B2B product. The async pipeline (RabbitMQ worker) is infrastructure. The check endpoint is what developers integrate.

**Design:**
```
POST /v1/check
Authorization: Bearer <api_key>
{
  "from": "alice@corp.com",
  "to": "external@gmail.com",
  "subject": "Q3 Financials",
  "message": "..."
}

→ 200 OK
{
  "verdict": "MEDIUM",
  "violations": ["unauthorized external financial disclosure"],
  "reasoning": "...",
  "remediated_body": "",
  "request_id": "uuid"
}
```

**Implementation path:**
1. Add `CheckEmail` handler to `tenant-service/cmd/api/handlers.go`
   - Validate API key via middleware (already exists)
   - Embed email via Mistral (already have `GeminiEmbedder` — a Mistral client)
   - Query RAG (policy chunks + history chunks via pgvector)
   - Run agent loop synchronously
   - Write audit log entry
   - Return structured JSON response
2. Add route `POST /v1/check` to `tenant-service/cmd/api/routes.go`
3. Wire `MistralAgent` into `Config` in `tenant-service/cmd/api/main.go` — currently only the embedder is wired, not the agent loop itself
4. Add check endpoint to front-end "Send Email" tab — instead of going to broker, send to `/v1/check` and show verdict inline before deciding to deliver

**Files to modify:**
- `tenant-service/cmd/api/handlers.go` — add `CheckEmail`
- `tenant-service/cmd/api/routes.go` — add route
- `tenant-service/cmd/api/main.go` — add `MistralAgent` to Config
- `front-end/cmd/web/templates/main.page.gohtml` — wire check to Send Email UI

---

## Phase 5: HIGH Priority Queue (P1)

**Why:** HIGH and MEDIUM currently behave identically in the review UI. HIGH should be visually distinct and optionally trigger an alert.

**Implementation:**
1. Add `priority` column to `quarantine` table:
   ```sql
   ALTER TABLE quarantine ADD COLUMN priority VARCHAR(10) NOT NULL DEFAULT 'normal'
     CHECK (priority IN ('normal', 'high'));
   ```
   Migration: `db-migrations/006_quarantine_priority.sql`

2. Update `InsertQuarantine` in `ai-compliance-service/data/models.go` to accept priority param
3. Update `runBlockedWorker` in `ai-compliance-service/cmd/api/pipeline.go` to pass `priority='high'`
4. Update `QueryQuarantine` in `tenant-service/data/models.go` to return priority field
5. Update `QuarantineEntry` struct to include `Priority` field
6. Update front-end quarantine table — red row highlight for HIGH, badge showing "HIGH PRIORITY"

---

## Phase 6: Tenant Compliance Settings (P1)

**Why:** Different organizations have different risk tolerances. A law firm may auto-block HIGH, a startup may just log everything.

**API:**
```
POST /v1/settings
Authorization: Bearer <api_key>
{
  "auto_deliver_low": true,      // default: true
  "quarantine_medium": true,     // default: true
  "quarantine_high": true,       // default: true (never auto-block)
  "retention_days": 90,          // quarantine retention
  "audit_retention_days": 365    // audit log retention
}
```

**Implementation:**
1. Add `tenant_settings` table:
   ```sql
   CREATE TABLE tenant_settings (
     tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
     auto_deliver_low BOOLEAN NOT NULL DEFAULT true,
     quarantine_medium BOOLEAN NOT NULL DEFAULT true,
     quarantine_high BOOLEAN NOT NULL DEFAULT true,
     retention_days INT NOT NULL DEFAULT 90,
     audit_retention_days INT NOT NULL DEFAULT 365,
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );
   ```
   Migration: `db-migrations/007_tenant_settings.sql`

2. Add `GetTenantSettings` + `UpsertTenantSettings` to `tenant-service/data/models.go`
3. Add `POST /v1/settings` and `GET /v1/settings` handlers to tenant-service
4. Thread settings into the compliance decision in `ai-compliance-service/cmd/api/pipeline.go`:
   - Before routing, fetch tenant settings (via direct DB query or HTTP call)
   - Apply routing rules based on settings

---

## Phase 7: Data Privacy Endpoints (P1)

**Why:** GDPR and HIPAA compliance is not optional for B2B enterprise customers. Required before any EU customer can sign up.

**Endpoints:**

### Right to Erasure (`DELETE /v1/data`)
```
DELETE /v1/data
Authorization: Bearer <api_key>
```
Cascades delete across all tenant-namespaced tables:
- `quarantine` WHERE `tenant_id = $1`
- `audit_log` WHERE `tenant_id = $1`
- `email_history_embeddings` WHERE `tenant_id = $1`
- `policy_embeddings` WHERE `tenant_id = $1`

Implementation: add `EraseAllTenantData(ctx, tenantID)` to `tenant-service/data/models.go`.

### Data Export (`GET /v1/export`)
```
GET /v1/export?format=json
Authorization: Bearer <api_key>
```
Returns all tenant data as a JSON archive. Streams from DB — don't load all into memory.

### Retention Job
Run as a goroutine in `tenant-service` or as a separate cron service:
```go
// Every hour: delete quarantine rows older than retention_days per tenant setting
DELETE FROM quarantine 
WHERE tenant_id = $1 AND created_at < NOW() - INTERVAL '1 day' * $2
```

---

## Phase 8: Encryption at Rest (P2)

**Why:** HIPAA requires encryption of PHI at rest. Quarantine bodies contain full email content.

**Approach:** Application-level AES-256-GCM encryption on `quarantine.body` before INSERT; decrypt on SELECT. Key stored in environment variable or AWS KMS.

**Files:**
- Add `encrypt(plaintext string, key []byte) ([]byte, error)` + `decrypt` helpers
- Update `InsertQuarantine` to encrypt body
- Update `QueryQuarantine` / `GetQuarantineByID` to decrypt body
- Add `ENCRYPTION_KEY` env var to docker-compose and Dockerfile

---

## Phase 9: Shield — SMTP Inbound Gateway (P3)

**Why:** The SMB product. Currently GoMailGuard only processes emails that are explicitly POSTed to it. Shield intercepts email at the DNS level — the SMB changes their MX record and all inbound mail flows through GoMailGuard first.

**Architecture:**
```
Internet → MX record → GoMailGuard SMTP gateway (port 25)
                              ↓
                    ai-compliance-service (scan)
                              ↓
                   ┌──────────┴──────────┐
                 CLEAN               MEDIUM/HIGH
                   ↓                      ↓
           Forward to tenant's       quarantine table
           actual mail server        (Shield review UI)
```

**Implementation:**
1. New service: `smtp-gateway` — listens on port 25, accepts SMTP, publishes to `email.ingest`
2. Use `github.com/emersion/go-smtp` for the SMTP server
3. Add `mx_record_verified` column to `tenants` table — tenant must prove domain ownership before activation
4. Shield dashboard: separate UI for inbound quarantine (same backend, different routing key label)
5. Onboarding flow: tenant adds `_gomail-verify.yourdomain.com TXT=<token>` for domain verification, then changes MX to `mx.gomail.io`

**Significant infra work** — estimate 2-3 weeks. This is the moat, not the MVP.

---

## Immediate Next Steps (do these first)

1. **Commit dirty binaries**: `git add ai-compliance-service/complianceApp tenant-service/tenantApp && git commit -m "build: update binaries"` — or add to `.gitignore` if binaries should not be committed
2. **Phase 4**: Synchronous check endpoint — this is the P0 feature that makes the product usable as a developer API
3. **Mistral DPA**: Reach out to Mistral to sign Data Processing Agreement before any production use with EU personal data

---

## Tech Debt / Known Bugs

| Issue | Location | Fix |
|-------|----------|-----|
| Binaries committed to git | `ai-compliance-service/complianceApp`, `tenant-service/tenantApp` | Add to `.gitignore` |
| `InsertQuarantine` violations param is JSON string not TEXT[] | `ai-compliance-service/data/models.go` | Either use TEXT[] with pq.Array or keep JSONB — currently mixed |
| No test for `CheckEmail` handler | `tenant-service/cmd/api/` | Add handler test once endpoint exists |
| `email.blocked` queue goes to same quarantine table as MEDIUM | `ai-compliance-service/cmd/api/pipeline.go` | Add priority field (Phase 5) |
| No retry/DLQ for AMQP nacks | `ai-compliance-service/cmd/api/pipeline.go` | Add dead-letter exchange binding |
| Front-end stores API key in localStorage | `front-end/cmd/web/templates/main.page.gohtml` | Acceptable for demo; use sessionStorage or secure cookie for prod |
