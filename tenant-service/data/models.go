package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	pgvector "github.com/pgvector/pgvector-go"
)

const dbTimeout = 3 * time.Second

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Models struct {
	db *sql.DB
}

func New(db *sql.DB) *Models {
	return &Models{db: db}
}

// CreateTenant inserts a new tenant and returns the created record.
func (m *Models) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `INSERT INTO tenants (name) VALUES ($1) RETURNING id, name, created_at`
	var t Tenant
	err := m.db.QueryRowContext(ctx, query, name).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GenerateAPIKey creates a random 32-byte key, stores its SHA-256 hash, and returns the raw key.
func (m *Models) GenerateAPIKey(ctx context.Context, tenantID, label string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	rawKey := hex.EncodeToString(b)

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	query := `INSERT INTO api_keys (tenant_id, key_hash, label) VALUES ($1, $2, $3)`
	if _, err := m.db.ExecContext(ctx, query, tenantID, keyHash, label); err != nil {
		return "", err
	}
	return rawKey, nil
}

// ValidateAPIKey checks a raw key against the stored hash and returns the tenant_id if valid.
func (m *Models) ValidateAPIKey(ctx context.Context, rawKey string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	query := `SELECT tenant_id FROM api_keys WHERE key_hash = $1 AND (expires_at IS NULL OR expires_at > NOW())`
	var tenantID string
	err := m.db.QueryRowContext(ctx, query, keyHash).Scan(&tenantID)
	return tenantID, err
}

// InsertPolicyEmbedding stores one chunk with its embedding vector.
func (m *Models) InsertPolicyEmbedding(ctx context.Context, tenantID, filename string, chunkIndex int, content string, embedding []float32) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `INSERT INTO policy_embeddings (tenant_id, source_filename, chunk_index, content, embedding) VALUES ($1, $2, $3, $4, $5)`
	_, err := m.db.ExecContext(ctx, query, tenantID, filename, chunkIndex, content, pgvector.NewVector(embedding))
	return err
}
