# GoMailGuard Phase 1: Infrastructure & Async Mail Pipeline

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the broker-service's direct HTTP mail call with an async RabbitMQ publish, add the email compliance exchange topology, add Redis and pgvector to the stack, and create the shared SQL schema migrations that all future services depend on.

**Architecture:** The broker-service no longer calls mail-service directly. It publishes a JSON `MailPayload` to the `email_events` RabbitMQ exchange with routing key `email.ingest`. The ai-compliance-service (Plan 3) will consume that queue, process the email, and call mail-service when approved. The new `email_events` topic exchange fans out to four downstream queues. Redis and pgvector-enabled Postgres are added to docker-compose as new infrastructure services.

**Tech Stack:** Go 1.22, `github.com/rabbitmq/amqp091-go v1.10.0` (existing), PostgreSQL 14 + pgvector extension, Redis 7, Docker Compose v3

---

## File Map

| Action | Path | Responsibility |
|---|---|---|
| Modify | `project/docker-compose.yml` | Switch postgres image → pgvector, add Redis service |
| Create | `db-migrations/001_tenants.sql` | Tenants + API keys schema |
| Create | `db-migrations/002_vector_tables.sql` | pgvector extension + embedding tables |
| Create | `db-migrations/003_audit_log.sql` | Audit trail table |
| Modify | `broker-service/event/event.go` | Add `declareEmailExchange` |
| Create | `broker-service/event/email_emitter.go` | `EmailEmitter` — publishes to `email_events` exchange |
| Modify | `broker-service/cmd/api/main.go` | Add `MailPublisher` interface + field to `Config`, wire `EmailEmitter` |
| Modify | `broker-service/cmd/api/handlers.go` | Update `sendMail` to use `MailPublisher` instead of HTTP |
| Create | `broker-service/cmd/api/handlers_test.go` | Unit tests for new `sendMail` behavior |

---

## Task 1: Switch Postgres to pgvector image and add Redis

**Files:**
- Modify: `project/docker-compose.yml`

- [ ] **Step 1: Update docker-compose.yml**

Replace the `postgres` service image and add the `redis` service. The pgvector image is a drop-in replacement for the official Postgres image — same configuration, with the `vector` extension pre-installed and available (but not yet enabled; that happens in the migration SQL).

```yaml
# project/docker-compose.yml
version: '3'

services:

  broker-service:
    build:
      context: ./../broker-service
      dockerfile: ./../broker-service/broker-service.dockerfile
    restart: always
    ports:
      - "8080:8080"
    deploy:
      mode: replicated
      replicas: 1
    environment:
      BROKER_URL: http://backend

  listener-service:
    build:
      context: ./../listener-service
      dockerfile: ./../listener-service/listener-service.dockerfile
    deploy:
      mode: replicated
      replicas: 1

  mail-service:
    build:
      context: ./../mail-service
      dockerfile: ./../mail-service/mail-service.dockerfile
    restart: always
    deploy:
      mode: replicated
      replicas: 1
    environment:
      MAIL_DOMAIN: localhost
      MAIL_HOST: mailhog
      MAIL_PORT: 1025
      MAIL_USERNAME: ""
      MAIL_PASSWORD: ""
      MAIL_ENCRYPTION: none
      FROM_NAME: "Minh Vu"
      FROM_ADDRESS: "minhvdq@example.com"

  logger-service:
    build:
      context: ./../logger-service
      dockerfile: ./../logger-service/logger-service.dockerfile
    restart: always
    deploy:
      mode: replicated
      replicas: 1

  authentication-service:
    build:
      context: ./../authentication-service
      dockerfile: ./../authentication-service/authentication-service.dockerfile
    restart: always
    ports:
      - "8081:80"
    deploy:
      mode: replicated
      replicas: 1
    environment:
      DSN: "host=postgres port=5432 user=postgres password=password dbname=users sslmode=disable timezone=UTC connect_timeout=5"

  postgres:
    image: 'pgvector/pgvector:pg14'
    ports:
      - "5432:5432"
    restart: always
    deploy:
      mode: replicated
      replicas: 1
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: users
    volumes:
      - ./db-data/postgres/:/var/lib/postgresql/data/

  redis:
    image: 'redis:7-alpine'
    ports:
      - "6379:6379"
    restart: always
    deploy:
      mode: replicated
      replicas: 1
    volumes:
      - ./db-data/redis/:/data

  mongo:
    image: 'mongo:4.2.16-bionic'
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_DATABASE: logs
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password
    volumes:
      - ./db-data/mongo/:/data/db

  mailhog:
    image: 'mailhog/mailhog'
    ports:
      - "1025:1025"
      - "8025:8025"

  rabbitmq:
    image: 'rabbitmq:3.13-alpine'
    ports:
      - "5672:5672"
    deploy:
      mode: replicated
      replicas: 1
    volumes:
      - ./db-data/rabbitmq/:/var/lib/rabbitmq/
```

- [ ] **Step 2: Verify containers start**

```bash
cd project && docker compose up -d postgres redis
```

Expected: both containers reach `healthy`/`running` state. Check with:
```bash
docker compose ps
```
Expected output includes `postgres` and `redis` with `Up` status.

- [ ] **Step 3: Verify pgvector extension is available**

```bash
docker compose exec postgres psql -U postgres -c "SELECT * FROM pg_available_extensions WHERE name = 'vector';"
```
Expected: one row returned with `name = vector`. The extension is available but not yet enabled (happens in Task 2).

- [ ] **Step 4: Commit**

```bash
git add project/docker-compose.yml
git commit -m "infra: switch postgres to pgvector image, add redis service"
```

---

## Task 2: Create SQL migration files

**Files:**
- Create: `db-migrations/001_tenants.sql`
- Create: `db-migrations/002_vector_tables.sql`
- Create: `db-migrations/003_audit_log.sql`

These are raw SQL files run in order. Plan 2 (tenant-service) will implement the migration runner. For now, create and verify them manually.

- [ ] **Step 1: Create db-migrations directory and 001_tenants.sql**

```bash
mkdir -p db-migrations
```

```sql
-- db-migrations/001_tenants.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS tenants (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID        NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key_hash    CHAR(64)    NOT NULL UNIQUE,
    label       VARCHAR(100),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys (tenant_id);
```

- [ ] **Step 2: Create 002_vector_tables.sql**

Gemini's `text-embedding-004` model outputs 768-dimensional vectors. The `ivfflat` index with `lists = 100` is appropriate for datasets up to ~1M rows.

```sql
-- db-migrations/002_vector_tables.sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS policy_embeddings (
    id              UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_filename VARCHAR(255),
    chunk_index     INT     NOT NULL,
    content         TEXT    NOT NULL,
    embedding       vector(768),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_policy_embeddings_tenant
    ON policy_embeddings (tenant_id);
CREATE INDEX IF NOT EXISTS idx_policy_embeddings_vec
    ON policy_embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE TABLE IF NOT EXISTS email_history_embeddings (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID    NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    content     TEXT    NOT NULL,
    embedding   vector(768),
    verdict     VARCHAR(10) NOT NULL CHECK (verdict IN ('CLEAN', 'LOW', 'MEDIUM', 'HIGH')),
    violations  TEXT[],
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_history_tenant
    ON email_history_embeddings (tenant_id);
CREATE INDEX IF NOT EXISTS idx_email_history_vec
    ON email_history_embeddings USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
```

- [ ] **Step 3: Create 003_audit_log.sql**

```sql
-- db-migrations/003_audit_log.sql
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
```

- [ ] **Step 4: Run migrations manually to verify SQL is valid**

```bash
cd project
docker compose exec postgres psql -U postgres -d users -f /dev/stdin < ../db-migrations/001_tenants.sql
docker compose exec postgres psql -U postgres -d users -f /dev/stdin < ../db-migrations/002_vector_tables.sql
docker compose exec postgres psql -U postgres -d users -f /dev/stdin < ../db-migrations/003_audit_log.sql
```

Expected: each command prints `CREATE TABLE` / `CREATE INDEX` / `CREATE EXTENSION` with no errors.

- [ ] **Step 5: Verify tables exist**

```bash
docker compose exec postgres psql -U postgres -d users -c "\dt"
```

Expected output includes: `tenants`, `api_keys`, `policy_embeddings`, `email_history_embeddings`, `audit_log`.

- [ ] **Step 6: Commit**

```bash
git add db-migrations/
git commit -m "infra: add SQL migrations for tenants, pgvector, and audit log"
```

---

## Task 3: Add email_events exchange to broker-service event package

**Files:**
- Modify: `broker-service/event/event.go`
- Create: `broker-service/event/email_emitter.go`

- [ ] **Step 1: Add declareEmailExchange to event.go**

Add the new function after the existing `declareRandomQueue` function. Do not remove any existing code.

```go
// broker-service/event/event.go
package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
}

func declareEmailExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"email_events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}
```

- [ ] **Step 2: Create email_emitter.go**

`EmailEmitter` publishes to `email_events`. It is separate from the existing `Emitter` (which publishes to `logs_topic`) so that concerns stay separated.

```go
// broker-service/event/email_emitter.go
package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type EmailEmitter struct {
	connection *amqp.Connection
}

func NewEmailEmitter(conn *amqp.Connection) (*EmailEmitter, error) {
	emitter := &EmailEmitter{connection: conn}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	if err := declareEmailExchange(ch); err != nil {
		return nil, err
	}

	return emitter, nil
}

func (e *EmailEmitter) Push(payload string, routingKey string) error {
	ch, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.Publish(
		"email_events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(payload),
		},
	)
}
```

- [ ] **Step 3: Verify broker-service compiles**

```bash
cd broker-service && go build ./...
```

Expected: no output (clean build).

- [ ] **Step 4: Commit**

```bash
git add broker-service/event/event.go broker-service/event/email_emitter.go
git commit -m "feat(broker): add email_events exchange and EmailEmitter"
```

---

## Task 4: Wire MailPublisher interface into Config

**Files:**
- Modify: `broker-service/cmd/api/main.go`

Introducing a `MailPublisher` interface makes `sendMail` unit-testable without a live RabbitMQ connection.

- [ ] **Step 1: Write the failing test first**

Create `broker-service/cmd/api/handlers_test.go`:

```go
// broker-service/cmd/api/handlers_test.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockMailPublisher struct {
	calls []struct {
		payload    string
		routingKey string
	}
	err error
}

func (m *mockMailPublisher) Push(payload, routingKey string) error {
	if m.err != nil {
		return m.err
	}
	m.calls = append(m.calls, struct {
		payload    string
		routingKey string
	}{payload, routingKey})
	return nil
}

func TestSendMailPublishesToEmailIngest(t *testing.T) {
	mock := &mockMailPublisher{}
	app := Config{MailPublisher: mock}

	msg := MailPayload{
		From:    "sender@college.edu",
		To:      "recipient@college.edu",
		Subject: "Test",
		Message: "Hello",
	}

	w := httptest.NewRecorder()
	app.sendMail(w, msg)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d", w.Code)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(mock.calls))
	}
	if mock.calls[0].routingKey != "email.ingest" {
		t.Errorf("expected routing key 'email.ingest', got %q", mock.calls[0].routingKey)
	}

	var published MailPayload
	if err := json.Unmarshal([]byte(mock.calls[0].payload), &published); err != nil {
		t.Fatalf("published payload is not valid JSON: %v", err)
	}
	if published.To != "recipient@college.edu" {
		t.Errorf("expected To=recipient@college.edu, got %q", published.To)
	}
}

func TestSendMailResponseBody(t *testing.T) {
	mock := &mockMailPublisher{}
	app := Config{MailPublisher: mock}

	w := httptest.NewRecorder()
	app.sendMail(w, MailPayload{From: "a@b.com", To: "c@d.com", Subject: "s", Message: "m"})

	var resp jsonResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error {
		t.Error("expected Error=false in response")
	}
	if resp.Message != "Message queued for compliance review" {
		t.Errorf("unexpected message: %q", resp.Message)
	}
}

func TestSendMailPublisherError(t *testing.T) {
	mock := &mockMailPublisher{err: fmt.Errorf("rabbitmq down")}
	app := Config{MailPublisher: mock}

	w := httptest.NewRecorder()
	app.sendMail(w, MailPayload{From: "a@b.com", To: "c@d.com", Subject: "s", Message: "m"})

	if w.Code == http.StatusAccepted {
		t.Error("expected non-202 when publisher errors")
	}
}
```

- [ ] **Step 2: Run the test — confirm it fails**

```bash
cd broker-service && go test ./cmd/api/ -v -run TestSendMail
```

Expected: compile error — `Config` has no field `MailPublisher`. This confirms the test is driving implementation.

- [ ] **Step 3: Add MailPublisher interface and field to main.go**

```go
// broker-service/cmd/api/main.go
package main

import (
	"broker/event"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "8080"

type MailPublisher interface {
	Push(payload string, routingKey string) error
}

type Config struct {
	Rabbit        *amqp.Connection
	MailPublisher MailPublisher
}

func main() {
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	emailEmitter, err := event.NewEmailEmitter(rabbitConn)
	if err != nil {
		log.Printf("failed to create email emitter: %v", err)
		os.Exit(1)
	}

	app := Config{
		Rabbit:        rabbitConn,
		MailPublisher: emailEmitter,
	}

	log.Printf("Starting broker on port %s\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ is not ready")
			counts++
		} else {
			connection = c
			log.Println("Connected to RabbitMQ")
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("Backing off")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
```

- [ ] **Step 4: Run the test — confirm it still fails (sendMail not updated yet)**

```bash
cd broker-service && go test ./cmd/api/ -v -run TestSendMail
```

Expected: compile success but test FAIL — `sendMail` still calls mail-service HTTP.

---

## Task 5: Update sendMail to publish to email.ingest

**Files:**
- Modify: `broker-service/cmd/api/handlers.go`

- [ ] **Step 1: Replace the sendMail function**

Find the existing `sendMail` function in `handlers.go` and replace it entirely. All other functions (`authenticate`, `logItem`, `logEventViaRabbit`, etc.) remain unchanged.

```go
func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	j, err := json.Marshal(msg)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if err := app.MailPublisher.Push(string(j), "email.ingest"); err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message queued for compliance review"

	app.writeJSON(w, http.StatusAccepted, payload)
}
```

- [ ] **Step 2: Run the tests — confirm all three pass**

```bash
cd broker-service && go test ./cmd/api/ -v -run TestSendMail
```

Expected:
```
--- PASS: TestSendMailPublishesToEmailIngest (0.00s)
--- PASS: TestSendMailResponseBody (0.00s)
--- PASS: TestSendMailPublisherError (0.00s)
PASS
```

- [ ] **Step 3: Verify full broker-service build is clean**

```bash
cd broker-service && go build ./...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add broker-service/cmd/api/main.go broker-service/cmd/api/handlers.go broker-service/cmd/api/handlers_test.go
git commit -m "feat(broker): async mail routing via email.ingest RabbitMQ queue"
```

---

## Verification: End-to-End Smoke Test

With docker compose running (`docker compose up -d` from `project/`), verify the full flow compiles and the broker responds correctly.

- [ ] **Step 1: Build and run broker-service locally against docker compose infra**

```bash
cd broker-service && go run ./cmd/api/
```

Expected: `Starting broker on port 8080` + `Connected to RabbitMQ`

- [ ] **Step 2: Send a mail request**

```bash
curl -s -X POST http://localhost:8080/handle \
  -H "Content-Type: application/json" \
  -d '{"Action":"mail","mail":{"from":"a@test.com","to":"b@test.com","subject":"hello","message":"world"}}' | jq .
```

Expected:
```json
{
  "error": false,
  "message": "Message queued for compliance review"
}
```

Note: the email is now in the `email_events` exchange with routing key `email.ingest`. It will not be delivered until Plan 3 (ai-compliance-service) consumes it.

- [ ] **Step 3: Final commit if any cleanup needed**

```bash
git add -p
git commit -m "chore: phase 1 infrastructure complete"
```

---

## What's Next

- **Plan 2:** `tenant-service` — org registration, API key management, policy document upload + pgvector embedding
- **Plan 3:** `ai-compliance-service` — Gemini agent loop, 5-phase detection pipeline, RAG, severity routing
- **Plan 4:** `quarantine-service` — queue consumer, held email storage, compliance officer review API
