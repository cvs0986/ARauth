package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// userRepository implements UserRepository for PostgreSQL
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, u *models.User) error {
	query := `
		INSERT INTO users (
			id, tenant_id, username, email, first_name, last_name,
			status, mfa_enabled, mfa_secret_encrypted, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	now := time.Now()
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	if u.Status == "" {
		u.Status = models.UserStatusActive
	}

	var metadataJSON []byte
	if u.Metadata != nil {
		metadataJSON, _ = json.Marshal(u.Metadata)
	}

	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.TenantID, u.Username, u.Email, u.FirstName, u.LastName,
		u.Status, u.MFAEnabled, u.MFASecretEncrypted,
		metadataJSON, // metadata as JSONB
		u.CreatedAt, u.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, tenant_id, username, email, first_name, last_name,
		       status, mfa_enabled, mfa_secret_encrypted, last_login_at,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	u := &models.User{}
	var firstName, lastName, mfaSecret sql.NullString
	var lastLoginAt, deletedAt sql.NullTime
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.Username, &u.Email,
		&firstName, &lastName, &u.Status, &u.MFAEnabled,
		&mfaSecret, &lastLoginAt, &metadataJSON, &u.CreatedAt,
		&u.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if firstName.Valid {
		u.FirstName = &firstName.String
	}
	if lastName.Valid {
		u.LastName = &lastName.String
	}
	if mfaSecret.Valid {
		u.MFASecretEncrypted = &mfaSecret.String
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	if deletedAt.Valid {
		u.DeletedAt = &deletedAt.Time
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &u.Metadata)
	}

	return u, nil
}

// GetByUsername retrieves a user by username and tenant ID
func (r *userRepository) GetByUsername(ctx context.Context, username string, tenantID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, tenant_id, username, email, first_name, last_name,
		       status, mfa_enabled, mfa_secret_encrypted, last_login_at,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND username = $2 AND deleted_at IS NULL
	`

	u := &models.User{}
	var firstName, lastName, mfaSecret sql.NullString
	var lastLoginAt, deletedAt sql.NullTime
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, username).Scan(
		&u.ID, &u.TenantID, &u.Username, &u.Email,
		&firstName, &lastName, &u.Status, &u.MFAEnabled,
		&mfaSecret, &lastLoginAt, &metadataJSON, &u.CreatedAt,
		&u.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	if firstName.Valid {
		u.FirstName = &firstName.String
	}
	if lastName.Valid {
		u.LastName = &lastName.String
	}
	if mfaSecret.Valid {
		u.MFASecretEncrypted = &mfaSecret.String
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	if deletedAt.Valid {
		u.DeletedAt = &deletedAt.Time
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &u.Metadata)
	}

	return u, nil
}

// GetByEmail retrieves a user by email and tenant ID
func (r *userRepository) GetByEmail(ctx context.Context, email string, tenantID uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, tenant_id, username, email, first_name, last_name,
		       status, mfa_enabled, mfa_secret_encrypted, last_login_at,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
	`

	u := &models.User{}
	var firstName, lastName, mfaSecret sql.NullString
	var lastLoginAt, deletedAt sql.NullTime
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Username, &u.Email,
		&firstName, &lastName, &u.Status, &u.MFAEnabled,
		&mfaSecret, &lastLoginAt, &metadataJSON, &u.CreatedAt,
		&u.UpdatedAt, &deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if firstName.Valid {
		u.FirstName = &firstName.String
	}
	if lastName.Valid {
		u.LastName = &lastName.String
	}
	if mfaSecret.Valid {
		u.MFASecretEncrypted = &mfaSecret.String
	}
	if lastLoginAt.Valid {
		u.LastLoginAt = &lastLoginAt.Time
	}
	if deletedAt.Valid {
		u.DeletedAt = &deletedAt.Time
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &u.Metadata)
	}

	return u, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, u *models.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, first_name = $4, last_name = $5,
		    status = $6, mfa_enabled = $7, mfa_secret_encrypted = $8,
		    last_login_at = $9, metadata = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL
	`

	u.UpdatedAt = time.Now()

	var metadataJSON []byte
	if u.Metadata != nil {
		metadataJSON, _ = json.Marshal(u.Metadata)
	}

	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.Username, u.Email, u.FirstName, u.LastName,
		u.Status, u.MFAEnabled, u.MFASecretEncrypted, u.LastLoginAt,
		metadataJSON, // metadata as JSONB
		u.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves a list of users with filters
func (r *userRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) ([]*user.User, error) {
	if filters == nil {
		filters = &interfaces.UserFilters{
			Page:     1,
			PageSize: 20,
		}
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}

	offset := (filters.Page - 1) * filters.PageSize

	query := `
		SELECT id, tenant_id, username, email, first_name, last_name,
		       status, mfa_enabled, mfa_secret_encrypted, last_login_at,
		       metadata, created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{tenantID}
	argPos := 2

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.Search != nil {
		query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)",
			argPos, argPos, argPos, argPos)
		searchPattern := "%" + *filters.Search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
		argPos += 4
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argPos) + " OFFSET $" + fmt.Sprintf("%d", argPos+1)
	args = append(args, filters.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

		var users []*models.User
	for rows.Next() {
		u := &models.User{}
		var firstName, lastName, mfaSecret sql.NullString
		var lastLoginAt, deletedAt sql.NullTime
		var metadataJSON []byte

		err := rows.Scan(
			&u.ID, &u.TenantID, &u.Username, &u.Email,
			&firstName, &lastName, &u.Status, &u.MFAEnabled,
			&mfaSecret, &lastLoginAt, &metadataJSON, &u.CreatedAt,
			&u.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if firstName.Valid {
			u.FirstName = &firstName.String
		}
		if lastName.Valid {
			u.LastName = &lastName.String
		}
		if mfaSecret.Valid {
			u.MFASecretEncrypted = &mfaSecret.String
		}
		if lastLoginAt.Valid {
			u.LastLoginAt = &lastLoginAt.Time
		}
		if deletedAt.Valid {
			u.DeletedAt = &deletedAt.Time
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &u.Metadata)
		}

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Count returns the total count of users matching filters
func (r *userRepository) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.UserFilters) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}
	argPos := 2

	if filters != nil {
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argPos)
			args = append(args, *filters.Status)
			argPos++
		}

		if filters.Search != nil {
			query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)",
				argPos, argPos, argPos, argPos)
			searchPattern := "%" + *filters.Search + "%"
			args = append(args, searchPattern, searchPattern, searchPattern, searchPattern)
		}
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

