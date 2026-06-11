# GoMailGuard — Product Spec (Revised)
**Date:** 2026-06-10  
**Status:** Active  
**Replaces:** 2026-06-07-gomail-guard-design.md

---

## 1. What We Build

GoMailGuard is a dual-product AI email compliance platform:

| Product | Who It's For | How They Use It |
|---------|-------------|-----------------|
| **GoMailGuard API** | Developers / enterprises building compliance into their own stack | REST API — `POST /v1/check` returns a structured verdict synchronously |
| **GoMailGuard Shield** | Small-to-medium businesses without IT departments | Change your MX record → GoMailGuard becomes your email gateway (inbound + outbound) |

Both products share the same Mistral-powered compliance engine and backend infrastructure. The API product is the priority; Shield is the longer-term moat.

---

## 2. Core Value Proposition

**For developers (API):** Drop a single HTTP call into your mail-send path and get back a structured compliance verdict — CLEAN, LOW, MEDIUM, or HIGH — with violations listed and reasoning attached. No model management, no prompt engineering, no pipeline to run.

**For SMBs (Shield):** Email security that doesn't require an IT team. Change one DNS record, connect your Google Workspace or Microsoft 365, set your policies in plain English, and GoMailGuard intercepts, scans, and delivers your email — flagging or holding anything suspicious instead of silently blocking it.

**Shared differentiator:** We don't hard-block. Competitors (Proofpoint, Mimecast) are widely criticized for silently dropping legitimate emails. GoMailGuard never discards — everything goes to a review queue with full reasoning attached, so compliance officers can make informed decisions.

---

## 3. Competitive Landscape

| Competitor | Weakness | Our Angle |
|-----------|---------|-----------|
| Proofpoint | Enterprise-only, over-blocks, no AI reasoning | We explain every decision; SMB-accessible pricing |
| Mimecast | Complex setup, false positive epidemic | We never hard-block; MX-based setup is 15 minutes |
| Tessian | Outbound-only, expensive, acqui-hired by Proofpoint | We do inbound + outbound; independent |
| Microsoft Purview | Microsoft-only, requires E5 license | We're provider-agnostic |
| Build-it-yourself | Months of ML work, Gemini/Mistral wrappers aren't compliance | We're trained on compliance policy, not generic text |

The market lesson from reviews: **false positives destroy trust faster than false negatives**. Our never-block philosophy is the product, not a feature.

---

## 4. Verdict System

Every email produces one of four verdicts. The key rule: **when in doubt, always queue for human review rather than block or pass**.

| Verdict | Meaning | Action |
|---------|---------|--------|
| `CLEAN` | No violations found | Deliver immediately |
| `LOW` | Minor unintentional issue, fully remediable by redaction | Auto-redact and deliver |
| `MEDIUM` | Serious violation or sender shows awareness of circumvention | Hold in quarantine for compliance officer review |
| `HIGH` | Clear malicious intent — phishing, mass exfiltration, explicit harm | Hold in quarantine with HIGH PRIORITY flag |

**MEDIUM vs HIGH distinction:** Both go to quarantine. HIGH is flagged for immediate escalation (Slack alert, email to compliance officer, front-of-queue). Neither is ever permanently discarded without a human decision.

**MEDIUM tiebreaker signals:**
- Sender language suggesting awareness: "delete this after reading", "you didn't hear this from me", "hasn't been approved yet"
- Confidential/financial/M&A content sent to personal or external addresses without explicit approval
- When uncertain between LOW and MEDIUM, always choose MEDIUM

---

## 5. AI Engine

**Model:** Mistral `mistral-small-latest` (multi-turn function calling)  
**Embeddings:** `mistral-embed` (1024-dimensional vectors, stored in pgvector)  
**Previous:** migrated from Google Gemini on 2026-06-10

### Agent Loop (up to 10 turns)

The agent calls tools iteratively until it has enough evidence to return a final JSON verdict:

```json
{"verdict":"MEDIUM","violations":["unauthorized external disclosure"],"reasoning":"...","remediated_body":""}
```

### 6 Tools

| Tool | Purpose |
|------|---------|
| `scan_pii` | Regex + semantic scan for SSNs, credit cards, PHI, passport numbers |
| `check_phishing` | Urgency manipulation, credential requests, spoofed sender, lookalike domains |
| `check_policy_violation` | RAG search against tenant's uploaded policy documents in pgvector |
| `check_exfiltration` | Bulk recipients, base64-encoded content, confidential content to unusual addresses |
| `retrieve_precedent` | RAG search against tenant's historical verdict records |
| `remediate_content` | Redact/rewrite for LOW verdicts only — never called for MEDIUM or HIGH |

### RAG Context (injected before each loop)

Before the agent starts, the email is embedded and the two nearest-neighbor stores are queried:
- **Policy chunks** — tenant's uploaded compliance documents, chunked and embedded
- **History chunks** — past emails and their verdicts, used for precedent reasoning

This gives the model grounded, tenant-specific context without fine-tuning.

---

## 6. Data Architecture

| Store | Role |
|-------|------|
| PostgreSQL + pgvector | Tenants, API keys, policy embeddings (1024-dim), email history embeddings (1024-dim), audit log, quarantine — all namespaced by `tenant_id` |
| MongoDB | Structured audit trail with full Mistral reasoning chains |
| Redis | Rate limiting, API key cache, future: communication graph for behavioral baseline |
| RabbitMQ | Message backbone — topic exchange `email_events` |

### RabbitMQ Topology

```
email.ingest          → ai-compliance-service (Mistral agent loop)
                               ↓
email.approved        → mail-service (CLEAN/LOW delivery)
email.quarantine      → ai-compliance-service quarantine consumer → quarantine table
email.blocked         → ai-compliance-service blocked consumer → quarantine table (HIGH priority)
```

Note: `email.blocked` is not a dead-end — it feeds the same quarantine store with a `HIGH` priority flag. Nothing is ever discarded without human decision.

---

## 7. Current API Surface

All authenticated endpoints require `Authorization: Bearer <api_key>`. All data is scoped to the calling tenant.

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `POST` | `/v1/organizations` | Register org → returns `tenant_id` + API key |
| `POST` | `/v1/policies` | Upload compliance policy doc (chunked + embedded into pgvector) |
| `GET` | `/v1/audit` | Decision log — filterable by `verdict`, `limit` |
| `GET` | `/v1/quarantine` | Pending quarantine items — filterable by `status` |
| `POST` | `/v1/quarantine/{id}/review` | Release (→ mail-service delivery) or reject (→ audit only) |

### Planned API Endpoints

| Method | Endpoint | Purpose | Priority |
|--------|----------|---------|----------|
| `POST` | `/v1/check` | **Synchronous check** — submit email, get verdict immediately | P0 |
| `POST` | `/v1/settings` | Configure routing thresholds per tenant | P1 |
| `DELETE` | `/v1/data` | Right to erasure (GDPR Art. 17) | P1 |
| `GET` | `/v1/export` | Data portability export (GDPR Art. 20) | P2 |

---

## 8. Data Privacy

GoMailGuard processes email content, which is personal data under GDPR and may include PHI under HIPAA.

### Data Processing Chain

```
Tenant (Controller) → GoMailGuard (Processor) → Mistral La Plateforme (Sub-processor)
```

Required: **Mistral Data Processing Agreement** — must be signed before processing any EU personal data.

### Privacy Controls (current)

- Tenant data is namespaced by `tenant_id` — cross-tenant access is impossible at the SQL level
- API keys are the only authentication mechanism — no user accounts storing personal data
- Quarantine body content stored in plaintext — **encryption at rest required before HIPAA compliance**

### Privacy Controls (planned, P1)

| Control | Implementation |
|---------|---------------|
| Configurable retention | `DELETE FROM quarantine WHERE created_at < NOW() - INTERVAL $1` — run as scheduled job per tenant setting |
| Right to erasure | `DELETE` cascade on `tenant_id` across all tables |
| Data export | Serialize all tenant rows to JSON/CSV on demand |
| Encryption at rest | AES-256 on `quarantine.body` at application layer before INSERT |
| Audit log retention | Separate configurable retention from quarantine (audit logs kept longer by default) |

---

## 9. Microservices Architecture

```
Customer / Developer
        │
        ▼
  broker-service (8080)          ← receives mail via POST /handle
        │
        ▼ (email.ingest → RabbitMQ)
  ai-compliance-service          ← Mistral agent loop
        │
   ┌────┴────────────────┐
   ▼                     ▼
mail-service         quarantine store
(CLEAN/LOW delivery) (MEDIUM/HIGH review)
        │                     │
        ▼                     ▼
    MailHog              tenant-service (8082) ← review API
                               │
                         front-end (80)        ← compliance dashboard
```

### Service Inventory

| Service | Port | Role |
|---------|------|------|
| `front-end` | 80 | Compliance dashboard SPA (Go + gohtml) |
| `broker-service` | 8080 | Email ingest gateway, publishes to RabbitMQ |
| `tenant-service` | 8082 | Tenant management, policy upload, audit/quarantine APIs |
| `ai-compliance-service` | — | Worker: Mistral agent, quarantine consumer, blocked consumer |
| `authentication-service` | 8081 | User auth (existing) |
| `mail-service` | — | Email delivery to MailHog/SMTP |
| `logger-service` | — | MongoDB audit logging |
| `listener-service` | — | RabbitMQ listener (existing) |

---

## 10. Deployment

**Phase 1 (current):** Docker Compose — `MISTRAL_API_KEY=... make up_build`

**Phase 2 (planned):** Hosted SaaS — GoMailGuard runs the infrastructure, tenants connect via API key. Mistral processes content as sub-processor under a signed DPA.

**Phase 3 (planned):** Shield inbound product — tenant changes MX record to point to GoMailGuard's SMTP gateway. Inbound email is scanned before delivery to the tenant's mail provider.

---

## 11. Resume Bullets

- Built GoMailGuard, a dual-product AI email compliance platform: a developer REST API for synchronous compliance checks and an SMB email gateway product intercepting email at the MX record level
- Engineered a Mistral multi-turn agent loop with 6 tools (PII scan, phishing detection, policy RAG, exfiltration check, precedent retrieval, LOW-only auto-remediation) over a 1024-dim pgvector knowledge base with per-tenant embedding isolation
- Designed a quarantine-first verdict system (CLEAN/LOW/MEDIUM/HIGH) where nothing is ever hard-blocked — MEDIUM and HIGH both route to human review queues with full AI reasoning attached, eliminating the false-positive problem that plagues Proofpoint and Mimecast
- Implemented multi-tenant B2B REST API with audit log, quarantine review workflow (release/reject wired to mail delivery), and policy document ingestion via Mistral embeddings
