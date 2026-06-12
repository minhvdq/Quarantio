package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

type OrgMember struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Role      string    `json:"role"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

type ReleaseRequest struct {
	ID             string     `json:"id"`
	QuarantineID   string     `json:"quarantine_id"`
	TenantID       string     `json:"tenant_id"`
	RequestedBy    string     `json:"requested_by"`
	RequesterEmail string     `json:"requester_email"`
	Note           string     `json:"note"`
	Status         string     `json:"status"`
	ReviewedBy     *string    `json:"reviewed_by,omitempty"`
	ReviewedAt     *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	EmailFrom      string     `json:"email_from"`
	Subject        string     `json:"subject"`
}

func (m *Models) CreateUser(ctx context.Context, email, password, firstName, lastName string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (email, password_hash, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, first_name, last_name, email_verified, created_at
	`
	var u User
	err = m.db.QueryRowContext(ctx, query, email, string(hash), firstName, lastName).
		Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.EmailVerified, &u.CreatedAt)
	return &u, err
}

func (m *Models) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `SELECT id, email, password_hash, first_name, last_name, email_verified, created_at FROM users WHERE email = $1`
	var u User
	err := m.db.QueryRowContext(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.EmailVerified, &u.CreatedAt)
	return &u, err
}

func (m *Models) GetUserByID(ctx context.Context, id string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `SELECT id, email, first_name, last_name, email_verified, created_at FROM users WHERE id = $1`
	var u User
	err := m.db.QueryRowContext(ctx, query, id).
		Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.EmailVerified, &u.CreatedAt)
	return &u, err
}

func (m *Models) CreateTenantWithDomain(ctx context.Context, name, domain string) (*Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `INSERT INTO tenants (name, domain) VALUES ($1, $2) RETURNING id, name, created_at`
	var t Tenant
	err := m.db.QueryRowContext(ctx, query, name, domain).Scan(&t.ID, &t.Name, &t.CreatedAt)
	return &t, err
}

func (m *Models) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `SELECT id, name, created_at FROM tenants WHERE domain = $1 AND domain_verified = true`
	var t Tenant
	err := m.db.QueryRowContext(ctx, query, domain).Scan(&t.ID, &t.Name, &t.CreatedAt)
	return &t, err
}

func (m *Models) CreateOrgMember(ctx context.Context, userID, tenantID, role string, invitedBy *string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	_, err := m.db.ExecContext(ctx,
		`INSERT INTO org_members (user_id, tenant_id, role, invited_by) VALUES ($1, $2, $3, $4)`,
		userID, tenantID, role, invitedBy,
	)
	return err
}

func (m *Models) GetOrgMember(ctx context.Context, userID, tenantID string) (*OrgMember, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		SELECT om.id, om.user_id, om.tenant_id, om.role, u.email, u.first_name, u.last_name, om.created_at
		FROM org_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.user_id = $1 AND om.tenant_id = $2
	`
	var om OrgMember
	err := m.db.QueryRowContext(ctx, query, userID, tenantID).
		Scan(&om.ID, &om.UserID, &om.TenantID, &om.Role, &om.Email, &om.FirstName, &om.LastName, &om.CreatedAt)
	return &om, err
}

// GetUserPrimaryTenant returns the first tenant the user belongs to (owner role preferred).
func (m *Models) GetUserPrimaryTenant(ctx context.Context, userID string) (*Tenant, string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		SELECT t.id, t.name, t.created_at, om.role
		FROM org_members om
		JOIN tenants t ON t.id = om.tenant_id
		WHERE om.user_id = $1
		ORDER BY CASE om.role WHEN 'owner' THEN 0 WHEN 'manager' THEN 1 WHEN 'monitor' THEN 2 ELSE 3 END
		LIMIT 1
	`
	var t Tenant
	var role string
	err := m.db.QueryRowContext(ctx, query, userID).Scan(&t.ID, &t.Name, &t.CreatedAt, &role)
	return &t, role, err
}

func (m *Models) ListOrgMembers(ctx context.Context, tenantID string) ([]OrgMember, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		SELECT om.id, om.user_id, om.tenant_id, om.role, u.email, u.first_name, u.last_name, om.created_at
		FROM org_members om
		JOIN users u ON u.id = om.user_id
		WHERE om.tenant_id = $1
		ORDER BY CASE om.role WHEN 'owner' THEN 0 WHEN 'manager' THEN 1 WHEN 'monitor' THEN 2 ELSE 3 END, om.created_at
	`
	rows, err := m.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []OrgMember
	for rows.Next() {
		var om OrgMember
		if err := rows.Scan(&om.ID, &om.UserID, &om.TenantID, &om.Role, &om.Email, &om.FirstName, &om.LastName, &om.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, om)
	}
	if members == nil {
		members = []OrgMember{}
	}
	return members, rows.Err()
}

func (m *Models) UpdateOrgMemberRole(ctx context.Context, memberID, tenantID, newRole string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	res, err := m.db.ExecContext(ctx,
		`UPDATE org_members SET role = $1 WHERE id = $2 AND tenant_id = $3 AND role != 'owner'`,
		newRole, memberID, tenantID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (m *Models) RemoveOrgMember(ctx context.Context, memberID, tenantID string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	res, err := m.db.ExecContext(ctx,
		`DELETE FROM org_members WHERE id = $1 AND tenant_id = $2 AND role != 'owner'`,
		memberID, tenantID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CreateSession generates a random refresh token, stores its hash, and returns the raw token.
func (m *Models) CreateSession(ctx context.Context, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	rawToken := hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	_, err := m.db.ExecContext(ctx,
		`INSERT INTO user_sessions (user_id, token_hash, expires_at) VALUES ($1, $2, NOW() + INTERVAL '7 days')`,
		userID, tokenHash,
	)
	return rawToken, err
}

func (m *Models) ValidateSession(ctx context.Context, rawToken string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	var userID string
	err := m.db.QueryRowContext(ctx,
		`SELECT user_id FROM user_sessions WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash,
	).Scan(&userID)
	return userID, err
}

func (m *Models) DeleteSession(ctx context.Context, rawToken string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	_, err := m.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (m *Models) CreateReleaseRequest(ctx context.Context, quarantineID, tenantID, userID, note string) (*ReleaseRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		INSERT INTO release_requests (quarantine_id, tenant_id, requested_by, note)
		VALUES ($1, $2, $3, $4)
		RETURNING id, quarantine_id, tenant_id, requested_by, note, status, created_at
	`
	var rr ReleaseRequest
	err := m.db.QueryRowContext(ctx, query, quarantineID, tenantID, userID, note).
		Scan(&rr.ID, &rr.QuarantineID, &rr.TenantID, &rr.RequestedBy, &rr.Note, &rr.Status, &rr.CreatedAt)
	return &rr, err
}

func (m *Models) ListReleaseRequests(ctx context.Context, tenantID, status string) ([]ReleaseRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	query := `
		SELECT rr.id, rr.quarantine_id, rr.tenant_id, rr.requested_by,
		       u.email, rr.note, rr.status, rr.reviewed_by, rr.reviewed_at, rr.created_at,
		       q.email_from, q.subject
		FROM release_requests rr
		JOIN users u ON u.id = rr.requested_by
		JOIN quarantine q ON q.id = rr.quarantine_id
		WHERE rr.tenant_id = $1 AND ($2 = '' OR rr.status = $2)
		ORDER BY rr.created_at DESC
	`
	rows, err := m.db.QueryContext(ctx, query, tenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReleaseRequest
	for rows.Next() {
		var rr ReleaseRequest
		if err := rows.Scan(
			&rr.ID, &rr.QuarantineID, &rr.TenantID, &rr.RequestedBy,
			&rr.RequesterEmail, &rr.Note, &rr.Status, &rr.ReviewedBy, &rr.ReviewedAt, &rr.CreatedAt,
			&rr.EmailFrom, &rr.Subject,
		); err != nil {
			return nil, err
		}
		results = append(results, rr)
	}
	if results == nil {
		results = []ReleaseRequest{}
	}
	return results, rows.Err()
}

func (m *Models) ActionReleaseRequest(ctx context.Context, requestID, tenantID, reviewerID, action string) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	res, err := m.db.ExecContext(ctx, `
		UPDATE release_requests
		SET status = $1, reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $3 AND tenant_id = $4 AND status = 'pending'
	`, action, reviewerID, requestID, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CheckPassword returns nil if the password matches the hash.
func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
