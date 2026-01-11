package oauthclient

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/arauth-identity/iam/storage/interfaces"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service implements OAuth2 client management with secure credential handling
type Service struct {
	repo interfaces.OAuthClientRepository
}

// NewService creates a new OAuth client service
func NewService(repo interfaces.OAuthClientRepository) ServiceInterface {
	return &Service{repo: repo}
}

// generateClientSecret generates a cryptographically secure client secret
// Returns 32 bytes of entropy, base64-encoded (43 characters)
// SECURITY: This secret is returned ONCE and never stored in plaintext
func generateClientSecret() (string, error) {
	bytes := make([]byte, 32) // 256 bits of entropy
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	// Base64 URL encoding for safe transmission
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashSecret hashes a client secret using bcrypt (cost 12)
// SECURITY: Secrets are NEVER stored in plaintext
func hashSecret(secret string) (string, error) {
	// Cost 12 = ~250ms on modern hardware (acceptable for client auth)
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}
	return string(hash), nil
}

// generateClientID generates a unique client ID
// Format: client_<32-char-hex>
func generateClientID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate client ID: %w", err)
	}
	return fmt.Sprintf("client_%s", hex.EncodeToString(bytes)), nil
}

// CreateClient creates a new OAuth2 client with generated credentials
// SECURITY: The secret is returned ONCE in the response and never again
func (s *Service) CreateClient(ctx context.Context, tenantID uuid.UUID, req *CreateClientRequest, createdBy uuid.UUID) (*CreateClientResponse, error) {
	// Generate client ID
	clientID, err := generateClientID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client ID: %w", err)
	}

	// Generate client secret (NEVER LOGGED)
	secret, err := generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Hash the secret for storage (bcrypt cost 12)
	secretHash, err := hashSecret(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash client secret: %w", err)
	}

	// Create client record
	client := &interfaces.OAuthClient{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Name:             req.Name,
		ClientID:         clientID,
		ClientSecretHash: secretHash, // NEVER plaintext
		Description:      &req.Description,
		RedirectURIs:     req.RedirectURIs,
		GrantTypes:       req.GrantTypes,
		Scopes:           req.Scopes,
		IsConfidential:   req.IsConfidential,
		IsActive:         true,
		CreatedBy:        &createdBy,
	}

	if err := s.repo.Create(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to create oauth client: %w", err)
	}

	// Return response with ONE-TIME secret
	// SECURITY: This is the ONLY time the plaintext secret is returned
	return &CreateClientResponse{
		ID:             client.ID,
		ClientID:       client.ClientID,
		ClientSecret:   secret, // ONE-TIME ONLY
		Name:           client.Name,
		Description:    req.Description,
		RedirectURIs:   client.RedirectURIs,
		GrantTypes:     client.GrantTypes,
		Scopes:         client.Scopes,
		IsConfidential: client.IsConfidential,
		CreatedAt:      client.CreatedAt,
	}, nil
}

// ListClients retrieves all clients for a tenant (WITHOUT secrets)
func (s *Service) ListClients(ctx context.Context, tenantID uuid.UUID) ([]*Client, error) {
	repoClients, err := s.repo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list oauth clients: %w", err)
	}

	// Map to client model (WITHOUT secret hash)
	clients := make([]*Client, len(repoClients))
	for i, rc := range repoClients {
		desc := ""
		if rc.Description != nil {
			desc = *rc.Description
		}
		clients[i] = &Client{
			ID:             rc.ID,
			ClientID:       rc.ClientID,
			Name:           rc.Name,
			Description:    desc,
			RedirectURIs:   rc.RedirectURIs,
			GrantTypes:     rc.GrantTypes,
			Scopes:         rc.Scopes,
			IsConfidential: rc.IsConfidential,
			IsActive:       rc.IsActive,
			CreatedAt:      rc.CreatedAt,
			UpdatedAt:      rc.UpdatedAt,
		}
	}

	return clients, nil
}

// GetClient retrieves a single client (WITHOUT secret)
// SECURITY: Tenant isolation enforced
func (s *Service) GetClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*Client, error) {
	repoClient, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client: %w", err)
	}

	// Enforce tenant isolation
	if repoClient.TenantID != tenantID {
		return nil, fmt.Errorf("oauth client does not belong to tenant")
	}

	desc := ""
	if repoClient.Description != nil {
		desc = *repoClient.Description
	}

	// Return client WITHOUT secret
	return &Client{
		ID:             repoClient.ID,
		ClientID:       repoClient.ClientID,
		Name:           repoClient.Name,
		Description:    desc,
		RedirectURIs:   repoClient.RedirectURIs,
		GrantTypes:     repoClient.GrantTypes,
		Scopes:         repoClient.Scopes,
		IsConfidential: repoClient.IsConfidential,
		IsActive:       repoClient.IsActive,
		CreatedAt:      repoClient.CreatedAt,
		UpdatedAt:      repoClient.UpdatedAt,
	}, nil
}

// RotateSecret generates a new secret and invalidates the old one
// SECURITY: The new secret is returned ONCE and never again
// TODO(Phase B4.1): Revoke all tokens issued with old secret
func (s *Service) RotateSecret(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*RotateSecretResponse, error) {
	// Get existing client
	repoClient, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth client: %w", err)
	}

	// Enforce tenant isolation
	if repoClient.TenantID != tenantID {
		return nil, fmt.Errorf("oauth client does not belong to tenant")
	}

	// Generate new secret (NEVER LOGGED)
	newSecret, err := generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new client secret: %w", err)
	}

	// Hash the new secret
	newSecretHash, err := hashSecret(newSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to hash new client secret: %w", err)
	}

	// Update client with new secret hash
	repoClient.ClientSecretHash = newSecretHash
	if err := s.repo.Update(ctx, repoClient); err != nil {
		return nil, fmt.Errorf("failed to update oauth client secret: %w", err)
	}

	// TODO(Phase B4.1): Revoke all refresh tokens issued with old secret
	// This requires RefreshTokenRepository.RevokeByClientID(clientID)
	// Deferred to Phase B4.1 to avoid partial implementation

	// Return response with ONE-TIME new secret
	return &RotateSecretResponse{
		ClientID:     repoClient.ClientID,
		ClientSecret: newSecret, // ONE-TIME ONLY
		RotatedAt:    repoClient.UpdatedAt,
	}, nil
}

// DeleteClient deletes a client
// SECURITY: Tenant isolation enforced
func (s *Service) DeleteClient(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// Get client to verify tenant ownership
	repoClient, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get oauth client: %w", err)
	}

	// Enforce tenant isolation
	if repoClient.TenantID != tenantID {
		return fmt.Errorf("oauth client does not belong to tenant")
	}

	// Delete client
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete oauth client: %w", err)
	}

	return nil
}
