package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OAuthClientRepository implements the OAuthClientRepository interface for PostgreSQL
type OAuthClientRepository struct {
	db *sql.DB
}

// NewOAuthClientRepository creates a new OAuth client repository
func NewOAuthClientRepository(db *sql.DB) interfaces.OAuthClientRepository {
	return &OAuthClientRepository{db: db}
}

// Create creates a new OAuth2 client
func (r *OAuthClientRepository) Create(ctx context.Context, client *interfaces.OAuthClient) error {
	query := `
		INSERT INTO oauth_clients (
			id, tenant_id, name, client_id, client_secret_hash,
			description, redirect_uris, grant_types, scopes,
			is_confidential, is_active, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		client.ID,
		client.TenantID,
		client.Name,
		client.ClientID,
		client.ClientSecretHash,
		client.Description,
		pq.Array(client.RedirectURIs),
		pq.Array(client.GrantTypes),
		pq.Array(client.Scopes),
		client.IsConfidential,
		client.IsActive,
		client.CreatedBy,
	).Scan(&client.CreatedAt, &client.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create oauth client: %w", err)
	}

	return nil
}

// GetByID retrieves a client by ID
func (r *OAuthClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*interfaces.OAuthClient, error) {
	query := `
		SELECT id, tenant_id, name, client_id, client_secret_hash,
		       description, redirect_uris, grant_types, scopes,
		       is_confidential, is_active, created_at, updated_at, created_by
		FROM oauth_clients
		WHERE id = $1
	`

	client := &interfaces.OAuthClient{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&client.ID,
		&client.TenantID,
		&client.Name,
		&client.ClientID,
		&client.ClientSecretHash,
		&client.Description,
		pq.Array(&client.RedirectURIs),
		pq.Array(&client.GrantTypes),
		pq.Array(&client.Scopes),
		&client.IsConfidential,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
		&client.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("oauth client not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client: %w", err)
	}

	return client, nil
}

// GetByClientID retrieves a client by client_id
func (r *OAuthClientRepository) GetByClientID(ctx context.Context, clientID string) (*interfaces.OAuthClient, error) {
	query := `
		SELECT id, tenant_id, name, client_id, client_secret_hash,
		       description, redirect_uris, grant_types, scopes,
		       is_confidential, is_active, created_at, updated_at, created_by
		FROM oauth_clients
		WHERE client_id = $1
	`

	client := &interfaces.OAuthClient{}
	err := r.db.QueryRowContext(ctx, query, clientID).Scan(
		&client.ID,
		&client.TenantID,
		&client.Name,
		&client.ClientID,
		&client.ClientSecretHash,
		&client.Description,
		pq.Array(&client.RedirectURIs),
		pq.Array(&client.GrantTypes),
		pq.Array(&client.Scopes),
		&client.IsConfidential,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
		&client.CreatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("oauth client not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client: %w", err)
	}

	return client, nil
}

// ListByTenant retrieves all clients for a tenant
func (r *OAuthClientRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*interfaces.OAuthClient, error) {
	query := `
		SELECT id, tenant_id, name, client_id, client_secret_hash,
		       description, redirect_uris, grant_types, scopes,
		       is_confidential, is_active, created_at, updated_at, created_by
		FROM oauth_clients
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list oauth clients: %w", err)
	}
	defer rows.Close()

	var clients []*interfaces.OAuthClient
	for rows.Next() {
		client := &interfaces.OAuthClient{}
		err := rows.Scan(
			&client.ID,
			&client.TenantID,
			&client.Name,
			&client.ClientID,
			&client.ClientSecretHash,
			&client.Description,
			pq.Array(&client.RedirectURIs),
			pq.Array(&client.GrantTypes),
			pq.Array(&client.Scopes),
			&client.IsConfidential,
			&client.IsActive,
			&client.CreatedAt,
			&client.UpdatedAt,
			&client.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan oauth client: %w", err)
		}
		clients = append(clients, client)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating oauth clients: %w", err)
	}

	return clients, nil
}

// Update updates an existing client
func (r *OAuthClientRepository) Update(ctx context.Context, client *interfaces.OAuthClient) error {
	query := `
		UPDATE oauth_clients
		SET name = $1, client_secret_hash = $2, description = $3,
		    redirect_uris = $4, grant_types = $5, scopes = $6,
		    is_confidential = $7, is_active = $8, updated_at = NOW()
		WHERE id = $9 AND tenant_id = $10
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		client.Name,
		client.ClientSecretHash,
		client.Description,
		pq.Array(client.RedirectURIs),
		pq.Array(client.GrantTypes),
		pq.Array(client.Scopes),
		client.IsConfidential,
		client.IsActive,
		client.ID,
		client.TenantID,
	).Scan(&client.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("oauth client not found or not owned by tenant")
	}
	if err != nil {
		return fmt.Errorf("failed to update oauth client: %w", err)
	}

	return nil
}

// Delete deletes a client by ID
func (r *OAuthClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM oauth_clients WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oauth client: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oauth client not found")
	}

	return nil
}
