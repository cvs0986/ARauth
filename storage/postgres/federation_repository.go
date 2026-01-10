package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/federation"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// identityProviderRepository implements IdentityProviderRepository
type identityProviderRepository struct {
	db *sql.DB
}

// NewIdentityProviderRepository creates a new identity provider repository
func NewIdentityProviderRepository(db *sql.DB) interfaces.IdentityProviderRepository {
	return &identityProviderRepository{db: db}
}

// Create creates a new identity provider
func (r *identityProviderRepository) Create(ctx context.Context, provider *federation.IdentityProvider) error {
	query := `
		INSERT INTO identity_providers (
			id, tenant_id, name, type, enabled, configuration, attribute_mapping,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	configJSON, err := json.Marshal(provider.Configuration)
	if err != nil {
		return err
	}

	var attrMappingJSON []byte
	if provider.AttributeMapping != nil {
		attrMappingJSON, err = json.Marshal(provider.AttributeMapping)
		if err != nil {
			return err
		}
	}

	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		provider.ID,
		provider.TenantID,
		provider.Name,
		provider.Type,
		provider.Enabled,
		configJSON,
		attrMappingJSON,
		provider.CreatedAt,
		provider.UpdatedAt,
	)

	return err
}

// GetByID retrieves an identity provider by ID
func (r *identityProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*federation.IdentityProvider, error) {
	query := `
		SELECT id, tenant_id, name, type, enabled, configuration, attribute_mapping,
		       created_at, updated_at, deleted_at
		FROM identity_providers
		WHERE id = $1 AND deleted_at IS NULL
	`

	var provider federation.IdentityProvider
	var configJSON, attrMappingJSON []byte

	var deletedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&provider.ID,
		&provider.TenantID,
		&provider.Name,
		&provider.Type,
		&provider.Enabled,
		&configJSON,
		&attrMappingJSON,
		&provider.CreatedAt,
		&provider.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("identity provider not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get identity provider: %w", err)
	}

	if deletedAt.Valid {
		provider.DeletedAt = &deletedAt.Time
	}

	// Unmarshal JSONB fields
	if err := json.Unmarshal(configJSON, &provider.Configuration); err != nil {
		return nil, err
	}

	if len(attrMappingJSON) > 0 {
		if err := json.Unmarshal(attrMappingJSON, &provider.AttributeMapping); err != nil {
			return nil, err
		}
	}

	return &provider, nil
}

// GetByTenantID retrieves all identity providers for a tenant
func (r *identityProviderRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*federation.IdentityProvider, error) {
	query := `
		SELECT id, tenant_id, name, type, enabled, configuration, attribute_mapping,
		       created_at, updated_at, deleted_at
		FROM identity_providers
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var providers []*federation.IdentityProvider
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var provider federation.IdentityProvider
		var configJSON, attrMappingJSON []byte
		var deletedAt sql.NullTime

		err := rows.Scan(
			&provider.ID,
			&provider.TenantID,
			&provider.Name,
			&provider.Type,
			&provider.Enabled,
			&configJSON,
			&attrMappingJSON,
			&provider.CreatedAt,
			&provider.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			provider.DeletedAt = &deletedAt.Time
		}

		// Unmarshal JSONB fields
		if err := json.Unmarshal(configJSON, &provider.Configuration); err != nil {
			return nil, err
		}

		if len(attrMappingJSON) > 0 {
			if err := json.Unmarshal(attrMappingJSON, &provider.AttributeMapping); err != nil {
				return nil, err
			}
		}

		providers = append(providers, &provider)
	}

	return providers, rows.Err()
}

// GetByName retrieves an identity provider by tenant ID and name
func (r *identityProviderRepository) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*federation.IdentityProvider, error) {
	query := `
		SELECT id, tenant_id, name, type, enabled, configuration, attribute_mapping,
		       created_at, updated_at, deleted_at
		FROM identity_providers
		WHERE tenant_id = $1 AND name = $2 AND deleted_at IS NULL
	`

	var provider federation.IdentityProvider
	var configJSON, attrMappingJSON []byte
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, tenantID, name).Scan(
		&provider.ID,
		&provider.TenantID,
		&provider.Name,
		&provider.Type,
		&provider.Enabled,
		&configJSON,
		&attrMappingJSON,
		&provider.CreatedAt,
		&provider.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("identity provider not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get identity provider: %w", err)
	}

	if deletedAt.Valid {
		provider.DeletedAt = &deletedAt.Time
	}

	// Unmarshal JSONB fields
	if err := json.Unmarshal(configJSON, &provider.Configuration); err != nil {
		return nil, err
	}

	if len(attrMappingJSON) > 0 {
		if err := json.Unmarshal(attrMappingJSON, &provider.AttributeMapping); err != nil {
			return nil, err
		}
	}

	return &provider, nil
}

// Update updates an existing identity provider
func (r *identityProviderRepository) Update(ctx context.Context, provider *federation.IdentityProvider) error {
	query := `
		UPDATE identity_providers
		SET name = $2, type = $3, enabled = $4, configuration = $5,
		    attribute_mapping = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	configJSON, err := json.Marshal(provider.Configuration)
	if err != nil {
		return err
	}

	var attrMappingJSON []byte
	if provider.AttributeMapping != nil {
		attrMappingJSON, err = json.Marshal(provider.AttributeMapping)
		if err != nil {
			return err
		}
	}

	provider.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		provider.ID,
		provider.Name,
		provider.Type,
		provider.Enabled,
		configJSON,
		attrMappingJSON,
		provider.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("identity provider not found")
	}

	return nil
}

// Delete soft deletes an identity provider
func (r *identityProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE identity_providers
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("identity provider not found")
	}

	return nil
}

// federatedIdentityRepository implements FederatedIdentityRepository
type federatedIdentityRepository struct {
	db *sql.DB
}

// NewFederatedIdentityRepository creates a new federated identity repository
func NewFederatedIdentityRepository(db *sql.DB) interfaces.FederatedIdentityRepository {
	return &federatedIdentityRepository{db: db}
}

// Create creates a new federated identity
func (r *federatedIdentityRepository) Create(ctx context.Context, identity *federation.FederatedIdentity) error {
	query := `
		INSERT INTO federated_identities (
			id, user_id, provider_id, external_id, attributes,
			is_primary, verified, verified_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`

	var attributesJSON []byte
	if identity.Attributes != nil {
		var err error
		attributesJSON, err = json.Marshal(identity.Attributes)
		if err != nil {
			return err
		}
	}

	now := time.Now()
	identity.CreatedAt = now
	identity.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		identity.ID,
		identity.UserID,
		identity.ProviderID,
		identity.ExternalID,
		attributesJSON,
		identity.IsPrimary,
		identity.Verified,
		identity.VerifiedAt,
		identity.CreatedAt,
		identity.UpdatedAt,
	)

	return err
}

// GetByID retrieves a federated identity by ID
func (r *federatedIdentityRepository) GetByID(ctx context.Context, id uuid.UUID) (*federation.FederatedIdentity, error) {
	query := `
		SELECT id, user_id, provider_id, external_id, attributes,
		       is_primary, verified, verified_at, created_at, updated_at
		FROM federated_identities
		WHERE id = $1
	`

	var identity federation.FederatedIdentity
	var attributesJSON []byte
	var verifiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.ProviderID,
		&identity.ExternalID,
		&attributesJSON,
		&identity.IsPrimary,
		&identity.Verified,
		&verifiedAt,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("identity provider not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get identity provider: %w", err)
	}

	if verifiedAt.Valid {
		identity.VerifiedAt = &verifiedAt.Time
	}

	// Unmarshal JSONB field
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &identity.Attributes); err != nil {
			return nil, err
		}
	}

	return &identity, nil
}

// GetByUserID retrieves all federated identities for a user
func (r *federatedIdentityRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*federation.FederatedIdentity, error) {
	query := `
		SELECT id, user_id, provider_id, external_id, attributes,
		       is_primary, verified, verified_at, created_at, updated_at
		FROM federated_identities
		WHERE user_id = $1
		ORDER BY is_primary DESC, created_at DESC
	`

	var identities []*federation.FederatedIdentity
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var identity federation.FederatedIdentity
		var attributesJSON []byte
		var verifiedAt sql.NullTime

		err := rows.Scan(
			&identity.ID,
			&identity.UserID,
			&identity.ProviderID,
			&identity.ExternalID,
			&attributesJSON,
			&identity.IsPrimary,
			&identity.Verified,
			&verifiedAt,
			&identity.CreatedAt,
			&identity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if verifiedAt.Valid {
			identity.VerifiedAt = &verifiedAt.Time
		}

		// Unmarshal JSONB field
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &identity.Attributes); err != nil {
				return nil, err
			}
		}

		identities = append(identities, &identity)
	}

	return identities, rows.Err()
}

// GetByProviderAndExternalID retrieves a federated identity by provider and external ID
func (r *federatedIdentityRepository) GetByProviderAndExternalID(ctx context.Context, providerID uuid.UUID, externalID string) (*federation.FederatedIdentity, error) {
	query := `
		SELECT id, user_id, provider_id, external_id, attributes,
		       is_primary, verified, verified_at, created_at, updated_at
		FROM federated_identities
		WHERE provider_id = $1 AND external_id = $2
	`

	var identity federation.FederatedIdentity
	var attributesJSON []byte
	var verifiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, providerID, externalID).Scan(
		&identity.ID,
		&identity.UserID,
		&identity.ProviderID,
		&identity.ExternalID,
		&attributesJSON,
		&identity.IsPrimary,
		&identity.Verified,
		&verifiedAt,
		&identity.CreatedAt,
		&identity.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("identity provider not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get identity provider: %w", err)
	}

	if verifiedAt.Valid {
		identity.VerifiedAt = &verifiedAt.Time
	}

	// Unmarshal JSONB field
	if len(attributesJSON) > 0 {
		if err := json.Unmarshal(attributesJSON, &identity.Attributes); err != nil {
			return nil, err
		}
	}

	return &identity, nil
}

// GetByProviderID retrieves all federated identities for a provider
func (r *federatedIdentityRepository) GetByProviderID(ctx context.Context, providerID uuid.UUID) ([]*federation.FederatedIdentity, error) {
	query := `
		SELECT id, user_id, provider_id, external_id, attributes,
		       is_primary, verified, verified_at, created_at, updated_at
		FROM federated_identities
		WHERE provider_id = $1
		ORDER BY created_at DESC
	`

	var identities []*federation.FederatedIdentity
	rows, err := r.db.QueryContext(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var identity federation.FederatedIdentity
		var attributesJSON []byte
		var verifiedAt sql.NullTime

		err := rows.Scan(
			&identity.ID,
			&identity.UserID,
			&identity.ProviderID,
			&identity.ExternalID,
			&attributesJSON,
			&identity.IsPrimary,
			&identity.Verified,
			&verifiedAt,
			&identity.CreatedAt,
			&identity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if verifiedAt.Valid {
			identity.VerifiedAt = &verifiedAt.Time
		}

		// Unmarshal JSONB field
		if len(attributesJSON) > 0 {
			if err := json.Unmarshal(attributesJSON, &identity.Attributes); err != nil {
				return nil, err
			}
		}

		identities = append(identities, &identity)
	}

	return identities, rows.Err()
}

// Update updates an existing federated identity
func (r *federatedIdentityRepository) Update(ctx context.Context, identity *federation.FederatedIdentity) error {
	query := `
		UPDATE federated_identities
		SET external_id = $2, attributes = $3, is_primary = $4,
		    verified = $5, verified_at = $6, updated_at = $7
		WHERE id = $1
	`

	var attributesJSON []byte
	if identity.Attributes != nil {
		var err error
		attributesJSON, err = json.Marshal(identity.Attributes)
		if err != nil {
			return err
		}
	}

	identity.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		identity.ID,
		identity.ExternalID,
		attributesJSON,
		identity.IsPrimary,
		identity.Verified,
		identity.VerifiedAt,
		identity.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("identity provider not found")
	}

	return nil
}

// Delete deletes a federated identity
func (r *federatedIdentityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM federated_identities WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("identity provider not found")
	}

	return nil
}

// SetPrimary sets a federated identity as primary (and unsets others for the user)
func (r *federatedIdentityRepository) SetPrimary(ctx context.Context, userID uuid.UUID, identityID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Unset all primary identities for this user
	_, err = tx.ExecContext(ctx, `
		UPDATE federated_identities
		SET is_primary = false, updated_at = NOW()
		WHERE user_id = $1 AND is_primary = true
	`, userID)
	if err != nil {
		return err
	}

	// Set the specified identity as primary
	_, err = tx.ExecContext(ctx, `
		UPDATE federated_identities
		SET is_primary = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, identityID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

