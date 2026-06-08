package data

import (
	"context"
	"database/sql"
	"time"
)

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

func (m *Models) CreateTenant(ctx context.Context, name string) (*Tenant, error)        { return nil, nil }
func (m *Models) GenerateAPIKey(ctx context.Context, tenantID, label string) (string, error) { return "", nil }
func (m *Models) ValidateAPIKey(ctx context.Context, rawKey string) (string, error)     { return "", nil }
func (m *Models) InsertPolicyEmbedding(ctx context.Context, tenantID, filename string, chunkIndex int, content string, embedding []float32) error { return nil }
