package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// InvitationRepository implements InvitationRepository interface
type InvitationRepository struct {
	db *sql.DB
}

// NewInvitationRepository creates a new invitation repository
func NewInvitationRepository(db *sql.DB) interfaces.InvitationRepository {
	return &InvitationRepository{db: db}
}

// Create creates a new invitation
func (r *InvitationRepository) Create(ctx context.Context, invitation *models.UserInvitation) error {
	query := `
		INSERT INTO user_invitations (
			id, tenant_id, email, invited_by, token_hash, expires_at,
			role_ids, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	if invitation.ID == uuid.Nil {
		invitation.ID = uuid.New()
	}
	if invitation.CreatedAt.IsZero() {
		invitation.CreatedAt = now
	}
	if invitation.UpdatedAt.IsZero() {
		invitation.UpdatedAt = now
	}

	// Convert role IDs to UUID array
	var roleIDsArray interface{}
	if len(invitation.RoleIDs) > 0 {
		roleIDsArray = pq.Array(invitation.RoleIDs)
	} else {
		roleIDsArray = pq.Array([]uuid.UUID{})
	}

	// Marshal metadata to JSONB
	var metadataJSON []byte
	if invitation.Metadata != nil {
		metadataJSON = []byte("{}") // TODO: Use proper JSON marshalling
	}

	_, err := r.db.ExecContext(ctx, query,
		invitation.ID,
		invitation.TenantID,
		invitation.Email,
		invitation.InvitedBy,
		invitation.TokenHash,
		invitation.ExpiresAt,
		roleIDsArray,
		metadataJSON,
		invitation.CreatedAt,
		invitation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

// GetByID retrieves an invitation by ID
func (r *InvitationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UserInvitation, error) {
	query := `
		SELECT id, tenant_id, email, invited_by, token_hash, expires_at,
		       accepted_at, accepted_by, role_ids, metadata, created_at, updated_at, deleted_at
		FROM user_invitations
		WHERE id = $1 AND deleted_at IS NULL
	`

	invitation := &models.UserInvitation{}
	var roleIDsArray pq.StringArray
	var acceptedAt, deletedAt sql.NullTime
	var acceptedBy sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&invitation.ID,
		&invitation.TenantID,
		&invitation.Email,
		&invitation.InvitedBy,
		&invitation.TokenHash,
		&invitation.ExpiresAt,
		&acceptedAt,
		&acceptedBy,
		&roleIDsArray,
		&metadataJSON,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// Convert role IDs array (UUID strings to UUIDs)
	if len(roleIDsArray) > 0 {
		invitation.RoleIDs = make([]uuid.UUID, len(roleIDsArray))
		for i, idStr := range roleIDsArray {
			if id, err := uuid.Parse(idStr); err == nil {
				invitation.RoleIDs[i] = id
			}
		}
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}
	if acceptedBy.Valid {
		if acceptedByUUID, err := uuid.Parse(acceptedBy.String); err == nil {
			invitation.AcceptedBy = &acceptedByUUID
		}
	}
	if deletedAt.Valid {
		invitation.DeletedAt = &deletedAt.Time
	}

	// TODO: Unmarshal metadata JSON

	return invitation, nil
}

// GetByTokenHash retrieves an invitation by token hash
func (r *InvitationRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.UserInvitation, error) {
	query := `
		SELECT id, tenant_id, email, invited_by, token_hash, expires_at,
		       accepted_at, accepted_by, role_ids, metadata, created_at, updated_at, deleted_at
		FROM user_invitations
		WHERE token_hash = $1 AND deleted_at IS NULL
	`

	invitation := &models.UserInvitation{}
	var roleIDsArray pq.StringArray
	var acceptedAt, deletedAt sql.NullTime
	var acceptedBy sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&invitation.ID,
		&invitation.TenantID,
		&invitation.Email,
		&invitation.InvitedBy,
		&invitation.TokenHash,
		&invitation.ExpiresAt,
		&acceptedAt,
		&acceptedBy,
		&roleIDsArray,
		&metadataJSON,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// Convert role IDs array (UUID strings to UUIDs)
	if len(roleIDsArray) > 0 {
		invitation.RoleIDs = make([]uuid.UUID, len(roleIDsArray))
		for i, idStr := range roleIDsArray {
			if id, err := uuid.Parse(idStr); err == nil {
				invitation.RoleIDs[i] = id
			}
		}
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}
	if acceptedBy.Valid {
		if acceptedByUUID, err := uuid.Parse(acceptedBy.String); err == nil {
			invitation.AcceptedBy = &acceptedByUUID
		}
	}
	if deletedAt.Valid {
		invitation.DeletedAt = &deletedAt.Time
	}

	return invitation, nil
}

// GetByEmail retrieves an invitation by email and tenant ID
func (r *InvitationRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.UserInvitation, error) {
	query := `
		SELECT id, tenant_id, email, invited_by, token_hash, expires_at,
		       accepted_at, accepted_by, role_ids, metadata, created_at, updated_at, deleted_at
		FROM user_invitations
		WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	invitation := &models.UserInvitation{}
	var roleIDsArray pq.StringArray
	var acceptedAt, deletedAt sql.NullTime
	var acceptedBy sql.NullString
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, tenantID, email).Scan(
		&invitation.ID,
		&invitation.TenantID,
		&invitation.Email,
		&invitation.InvitedBy,
		&invitation.TokenHash,
		&invitation.ExpiresAt,
		&acceptedAt,
		&acceptedBy,
		&roleIDsArray,
		&metadataJSON,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
		&deletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	// Convert role IDs array (UUID strings to UUIDs)
	if len(roleIDsArray) > 0 {
		invitation.RoleIDs = make([]uuid.UUID, len(roleIDsArray))
		for i, idStr := range roleIDsArray {
			if id, err := uuid.Parse(idStr); err == nil {
				invitation.RoleIDs[i] = id
			}
		}
	}

	if acceptedAt.Valid {
		invitation.AcceptedAt = &acceptedAt.Time
	}
	if acceptedBy.Valid {
		if acceptedByUUID, err := uuid.Parse(acceptedBy.String); err == nil {
			invitation.AcceptedBy = &acceptedByUUID
		}
	}
	if deletedAt.Valid {
		invitation.DeletedAt = &deletedAt.Time
	}

	return invitation, nil
}

// List lists invitations for a tenant
func (r *InvitationRepository) List(ctx context.Context, tenantID uuid.UUID, filters *interfaces.InvitationFilters) ([]*models.UserInvitation, error) {
	query := `
		SELECT id, tenant_id, email, invited_by, token_hash, expires_at,
		       accepted_at, accepted_by, role_ids, metadata, created_at, updated_at, deleted_at
		FROM user_invitations
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{tenantID}
	argIndex := 2

	// Apply filters
	if filters != nil {
		if filters.Email != "" {
			query += fmt.Sprintf(" AND email = $%d", argIndex)
			args = append(args, filters.Email)
			argIndex++
		}
		if filters.InvitedBy != nil {
			query += fmt.Sprintf(" AND invited_by = $%d", argIndex)
			args = append(args, *filters.InvitedBy)
			argIndex++
		}
		if filters.Status != "" {
			switch filters.Status {
			case "pending":
				query += " AND accepted_at IS NULL AND expires_at > CURRENT_TIMESTAMP"
			case "accepted":
				query += " AND accepted_at IS NOT NULL"
			case "expired":
				query += " AND accepted_at IS NULL AND expires_at <= CURRENT_TIMESTAMP"
			}
		}
	}

	query += " ORDER BY created_at DESC"

	// Apply pagination
	if filters != nil && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filters.PageSize, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query invitations: %w", err)
	}
	defer rows.Close()

	var invitations []*models.UserInvitation
	for rows.Next() {
		invitation := &models.UserInvitation{}
		var roleIDsArray pq.StringArray
		var acceptedAt, deletedAt sql.NullTime
		var acceptedBy sql.NullString
		var metadataJSON []byte

		err := rows.Scan(
			&invitation.ID,
			&invitation.TenantID,
			&invitation.Email,
			&invitation.InvitedBy,
			&invitation.TokenHash,
			&invitation.ExpiresAt,
			&acceptedAt,
			&acceptedBy,
			&roleIDsArray,
			&metadataJSON,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
			&deletedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}

		// Convert role IDs array (UUID strings to UUIDs)
		if len(roleIDsArray) > 0 {
			invitation.RoleIDs = make([]uuid.UUID, len(roleIDsArray))
			for i, idStr := range roleIDsArray {
				if id, err := uuid.Parse(idStr); err == nil {
					invitation.RoleIDs[i] = id
				}
			}
		}

		if acceptedAt.Valid {
			invitation.AcceptedAt = &acceptedAt.Time
		}
		if acceptedBy.Valid {
			if acceptedByUUID, err := uuid.Parse(acceptedBy.String); err == nil {
				invitation.AcceptedBy = &acceptedByUUID
			}
		}
		if deletedAt.Valid {
			invitation.DeletedAt = &deletedAt.Time
		}

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

// Count counts invitations for a tenant
func (r *InvitationRepository) Count(ctx context.Context, tenantID uuid.UUID, filters *interfaces.InvitationFilters) (int, error) {
	query := `SELECT COUNT(*) FROM user_invitations WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []interface{}{tenantID}
	argIndex := 2

	// Apply filters
	if filters != nil {
		if filters.Email != "" {
			query += fmt.Sprintf(" AND email = $%d", argIndex)
			args = append(args, filters.Email)
			argIndex++
		}
		if filters.InvitedBy != nil {
			query += fmt.Sprintf(" AND invited_by = $%d", argIndex)
			args = append(args, *filters.InvitedBy)
			argIndex++
		}
		if filters.Status != "" {
			switch filters.Status {
			case "pending":
				query += " AND accepted_at IS NULL AND expires_at > CURRENT_TIMESTAMP"
			case "accepted":
				query += " AND accepted_at IS NOT NULL"
			case "expired":
				query += " AND accepted_at IS NULL AND expires_at <= CURRENT_TIMESTAMP"
			}
		}
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count invitations: %w", err)
	}

	return count, nil
}

// Update updates an invitation
func (r *InvitationRepository) Update(ctx context.Context, invitation *models.UserInvitation) error {
	query := `
		UPDATE user_invitations
		SET email = $1, expires_at = $2, accepted_at = $3, accepted_by = $4,
		    role_ids = $5, metadata = $6, updated_at = $7
		WHERE id = $8 AND deleted_at IS NULL
	`

	invitation.UpdatedAt = time.Now()

	// Convert role IDs to UUID array
	var roleIDsArray interface{}
	if len(invitation.RoleIDs) > 0 {
		roleIDsArray = pq.Array(invitation.RoleIDs)
	} else {
		roleIDsArray = pq.Array([]uuid.UUID{})
	}

	// Marshal metadata to JSONB
	var metadataJSON []byte
	if invitation.Metadata != nil {
		metadataJSON = []byte("{}") // TODO: Use proper JSON marshalling
	}

	var acceptedAt interface{}
	if invitation.AcceptedAt != nil {
		acceptedAt = *invitation.AcceptedAt
	}

	var acceptedBy interface{}
	if invitation.AcceptedBy != nil {
		acceptedBy = *invitation.AcceptedBy
	}

	result, err := r.db.ExecContext(ctx, query,
		invitation.Email,
		invitation.ExpiresAt,
		acceptedAt,
		acceptedBy,
		roleIDsArray,
		metadataJSON,
		invitation.UpdatedAt,
		invitation.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already deleted")
	}

	return nil
}

// Delete soft-deletes an invitation
func (r *InvitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE user_invitations
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation not found or already deleted")
	}

	return nil
}

