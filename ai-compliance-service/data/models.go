// ai-compliance-service/data/models.go
package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	pgvector "github.com/pgvector/pgvector-go"
)

const dbTimeout = 5 * time.Second

type Models struct {
	db *sql.DB
}

func New(db *sql.DB) *Models {
	return &Models{db: db}
}

// ChunkRow holds one row returned by a pgvector similarity query.
type ChunkRow struct {
	Content string
	Source  string
}

// InsertAuditLog writes one compliance decision to the audit_log table.
func (m *Models) InsertAuditLog(ctx context.Context, tenantID, emailFrom, emailSubject, verdict, reasoning, action string, emailTo, violations []string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	vb, _ := json.Marshal(violations)
	violationsJSON := string(vb)

	query := `
		INSERT INTO audit_log
			(tenant_id, email_from, email_to, email_subject, verdict, violations, gemini_reasoning, action_taken)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
	`
	_, err := m.db.ExecContext(ctx, query,
		tenantID,
		emailFrom,
		emailTo,
		emailSubject,
		verdict,
		violationsJSON,
		reasoning,
		action,
	)
	return err
}

// InsertEmailHistory stores an email embedding for future precedent RAG queries.
func (m *Models) InsertEmailHistory(ctx context.Context, tenantID, content string, embedding []float32, verdict, violations string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		INSERT INTO email_history_embeddings (tenant_id, content, embedding, verdict, violations)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := m.db.ExecContext(ctx, query,
		tenantID,
		content,
		pgvector.NewVector(embedding),
		verdict,
		violations,
	)
	return err
}

// QueryPolicyChunks returns up to limit policy chunks nearest to the given embedding.
func (m *Models) QueryPolicyChunks(ctx context.Context, tenantID string, embedding []float32, limit int) ([]ChunkRow, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		SELECT content, COALESCE(source_filename, 'policy') AS source
		FROM policy_embeddings
		WHERE tenant_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3
	`
	rows, err := m.db.QueryContext(ctx, query, tenantID, pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []ChunkRow
	for rows.Next() {
		var c ChunkRow
		if err := rows.Scan(&c.Content, &c.Source); err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}

// QueryHistoryChunks returns up to limit historical email entries nearest to the given embedding.
func (m *Models) QueryHistoryChunks(ctx context.Context, tenantID string, embedding []float32, limit int) ([]ChunkRow, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		SELECT content, verdict AS source
		FROM email_history_embeddings
		WHERE tenant_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3
	`
	rows, err := m.db.QueryContext(ctx, query, tenantID, pgvector.NewVector(embedding), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chunks []ChunkRow
	for rows.Next() {
		var c ChunkRow
		if err := rows.Scan(&c.Content, &c.Source); err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}
