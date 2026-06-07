# GoMailGuard — Design Spec
**Date:** 2026-06-07  
**Status:** Approved  
**License:** AGPL v3

---

## 1. Product Overview

GoMailGuard is an open-source, self-hosted B2B email compliance platform built on top of the GoMail microservices stack. Organizations (universities, hospitals, enterprises) deploy the full stack inside their own infrastructure. Every outbound email is intercepted, analyzed by a Gemini AI agent loop, and routed based on a tiered severity verdict before delivery.

**Core value proposition:** Enterprise-grade AI compliance enforcement that lives entirely inside the customer's firewall. No raw email content ever leaves their environment.

**Business model (open core):**
- Free: self-hosted, full pipeline, community support
- Paid: hosted cloud version, premium SLA, cross-tenant threat intelligence feed

---

## 2. The Trust Model

Customers hand GoMailGuard access to their email stream. The architectural response to this trust requirement:

1. **On-premise deployment only** — ships as Docker Compose (small orgs) and Helm chart (Kubernetes/production). Customer owns all infrastructure.
2. **AGPL v3 license** — full source available for audit. No black-box components.
3. **Zero data egress by default** — no email content, metadata, or PII ever leaves the customer's environment unless they explicitly opt into the anonymized threat signal feed.
4. **Tenant isolation** — all data namespaced by `tenant_id`; orgs within a shared deployment cannot access each other's data.

---

## 3. Five-Phase Detection Pipeline

### Phase 1 — Ingestion & Deconstruction
- Hook into mail flow via Microsoft Graph API journaling or Google Workspace Gmail API webhooks
- Raw MIME message published immediately to `email.ingest` RabbitMQ queue (zero message loss guarantee)
- Deconstruction worker splits payload into three streams:
  - **Metadata**: headers, SPF/DKIM/DMARC status, hop counts, sender IP
  - **Body content**: raw text, parsed HTML DOM structure
  - **Artifacts**: extracted URLs, file attachments

### Phase 2 — Contextual Enrichment & Graph Lookup
- Query Redis Communication Graph to establish behavioral baseline before any LLM inference:
  - Have sender and recipient interacted in the last 90 days?
  - Is sender IP outside their normal geographic cluster?
  - Does sender display name match a known executive but use an unmapped external domain?
- Enrichment output is attached to the message payload before Phase 3

### Phase 3 — Parallel Feature Analysis (Pluggable Ensemble)
Three analyzers run concurrently, results aggregated into a feature vector:

- **Linguistic vectorizer**: Gemini NLU classifies semantic intent — financial coercion, urgent credential requests, impersonation language
- **LLM fingerprinter**: Analyzes text perplexity and burstiness to score probability of synthetically generated (attacker LLM) content
- **Heuristic engine**: Parses raw HTML for hidden elements, unrendered markdown, obfuscated redirect scripts, missing legal disclaimers

All three analyzers implement a standard `Analyzer` interface — **pluggable modules** that the community can extend without modifying core code.

### Phase 4 — Dynamic Link & Attachment Detonation (threshold-gated)
- Only triggered when Phase 3 composite score exceeds anomaly threshold (avoids unnecessary overhead on clean mail)
- Extracts all embedded URLs and attachments
- Routes to isolated sandbox environment
- Traces redirect chains, detects lookalike login pages, records JavaScript telemetry
- Scoped as a pluggable module; ships disabled by default, enabled via config flag

### Phase 5 — Ensemble Scoring & Mitigation
- Aggregates graph enrichment, NLU output, heuristic flags, and sandbox results into a composite risk vector
- Severity verdict:

| Severity | Condition | Action |
|---|---|---|
| `CLEAN` | No signals | → `email.approved` → mail-service delivers |
| `LOW` | Minor flags, auto-remediable | → Gemini remediates content → `email.approved` |
| `MEDIUM` | Ambiguous risk | → `email.quarantine` → compliance officer reviews |
| `HIGH` | High-confidence threat | → `email.blocked` → dead-lettered + audited immediately |

---

## 4. Gemini Agent Loop

The `ai-compliance-service` runs a multi-turn agentic reasoning loop using Gemini function calling. Gemini calls tools iteratively until it has sufficient evidence to return a final verdict.

**6 tools:**

| Tool | Purpose |
|---|---|
| `scan_pii` | Detect SSNs, credit card numbers, PHI, passport numbers (regex + semantic) |
| `check_phishing` | Analyze URLs, sender spoofing, urgency manipulation language |
| `check_policy_violation` | RAG search against tenant's policy documents in pgvector |
| `check_exfiltration` | Flag unusual bulk recipient lists, suspicious attachment signatures |
| `retrieve_precedent` | RAG search against tenant's historical approved/flagged emails |
| `remediate_content` | Redact PII or rewrite flagged sections for LOW severity cases |

**RAG context injection:**
Before the loop starts, email content is embedded and queried against pgvector for:
- Matching policy document chunks (tenant's uploaded compliance rules)
- Similar historical emails and their past verdicts

Retrieved context is injected into Gemini's system prompt, giving it grounded, tenant-specific knowledge before tool calls begin.

---

## 5. Data Architecture

| Store | Role |
|---|---|
| PostgreSQL + pgvector | Tenant registry, API keys, policy embeddings, historical email embeddings (all namespaced by `tenant_id`) |
| Redis | Communication Graph — sender/recipient interaction history, IP geocluster baselines, executive name registry |
| MongoDB | Full audit trail — every decision record: headers, verdict, tool outputs, Gemini reasoning chain, compliance officer actions |
| RabbitMQ | Message backbone — expanded queue topology (see below) |

**RabbitMQ topology:**
```
email.ingest         → ai-compliance-service (Phase 1: deconstruction)
                              ↓
email.compliance     → ai-compliance-service (Phase 2–5: Gemini loop)
                              ↓
email.approved       → mail-service (existing)
email.quarantine     → quarantine-service
email.blocked        → logger-service (audit)
email.signals        → optional central threat intel (anonymized only, opt-in)
```

---

## 6. New Microservices

**`tenant-service`**
- Org registration and API key management
- Policy document ingestion: accepts PDF/text uploads, chunks and embeds into pgvector
- Backed by PostgreSQL

**`ai-compliance-service`**
- Consumes `email.ingest` (Phase 1: MIME deconstruction, publishes enriched payload to `email.compliance`)
- Consumes `email.compliance` (Phase 2–5: graph lookup, parallel analysis, Gemini agent loop)
- Hosts the Gemini agent loop and all 6 tools
- Queries pgvector (RAG) and Redis (graph) for context
- Routes to downstream queues based on severity verdict
- Writes decision records to MongoDB

**`quarantine-service`**
- Consumes `email.quarantine` queue
- Stores held emails with full metadata
- Exposes REST API for compliance officer review
- On approval: re-publishes to `email.approved`
- On rejection: writes final audit record to MongoDB

---

## 7. B2B API

All endpoints require `Authorization: Bearer <api_key>`. All responses scoped to the calling tenant.

| Endpoint | Purpose |
|---|---|
| `POST /v1/organizations` | Register org, returns `tenant_id` + API key |
| `POST /v1/policies` | Upload compliance policy doc (auto-embedded into pgvector) |
| `GET /v1/quarantine` | List emails pending compliance review |
| `POST /v1/quarantine/{id}/approve` | Release held email to delivery |
| `POST /v1/quarantine/{id}/reject` | Hard block + audit record |
| `GET /v1/audit` | Full decision log with Gemini reasoning chain |
| `GET /v1/dashboard` | Risk stats: volume, severity breakdown, top violation types |
| `POST /v1/send` | Direct transactional send (bypasses journaling) |

Every response includes a `request_id` traceable to a full audit record.

**Onboarding flow (15 minutes for an IT admin):**
1. `POST /v1/organizations` → receive API key
2. Configure Microsoft 365 or Google Workspace to journal outbound mail to GoMailGuard webhook
3. `POST /v1/policies` → upload compliance policy documents
4. Emails begin flowing through the pipeline automatically

---

## 8. Deployment Model

**Self-hosted, on-premise first.**

```
Customer's Environment
├── tenant-service          (new)
├── ai-compliance-service   (new)
├── quarantine-service      (new)
├── broker-service          (existing, extended)
├── authentication-service  (existing)
├── logger-service          (existing, extended)
├── mail-service            (existing)
├── listener-service        (existing)
├── postgres + pgvector     (existing, extended)
├── mongodb                 (existing, extended)
├── redis                   (new)
└── rabbitmq                (existing, extended)
```

**Customer `.env` configuration:**
```
GEMINI_API_KEY=...
MAIL_PROVIDER=microsoft365   # or google_workspace
MICROSOFT_TENANT_ID=...
MICROSOFT_CLIENT_SECRET=...
POSTGRES_DSN=...
REDIS_URL=...
PHASE4_SANDBOX_ENABLED=false  # opt-in
THREAT_SIGNALS_ENABLED=false  # opt-in
```

Ships with:
- `docker-compose.yml` for development and small orgs
- Helm chart for Kubernetes production deployments
- One-command quickstart: `docker compose up`

---

## 9. Open Source Strategy

- **License:** AGPL v3 — all deployments (including SaaS wrappers) must open source modifications
- **Monorepo:** Single GitHub repository with clear per-service directories
- **Plugin interface:** Phase 3 analyzers implement a standard `Analyzer` interface; community contributions add new detectors without touching core
- **README:** Architecture diagram, quickstart, and integration guides for Microsoft 365 and Google Workspace
- **Roadmap items (future):** Phase 4 sandbox module, inbound email support, Slack/Teams alert integration, SIEM export

---

## 10. Resume Bullets (target)

- Architected GoMailGuard, an open-source AGPL B2B email compliance platform; engineered a 5-phase detection pipeline (ingestion, graph enrichment, parallel NLU ensemble, sandbox detonation, mitigation) deployed on-premise inside customer infrastructure
- Engineered an agentic Gemini reasoning loop with 6 tools (PII scan, phishing detection, policy RAG, exfiltration check, precedent retrieval, auto-remediation) over a pgvector knowledge base with per-tenant embedding isolation
- Designed a multi-tenant B2B REST API enabling 15-minute onboarding for Microsoft 365 and Google Workspace organizations with full audit trail, quarantine review workflow, and zero data egress guarantees
