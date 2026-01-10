package scim

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// TokenService provides SCIM token management
type TokenService struct {
	tokenRepo interfaces.SCIMTokenRepository
}

// NewTokenService creates a new SCIM token service
func NewTokenService(tokenRepo interfaces.SCIMTokenRepository) TokenServiceInterface {
	return &TokenService{
		tokenRepo: tokenRepo,
	}
}

// generateToken generates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashToken hashes a token using bcrypt
func hashToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}
	return string(hash), nil
}

// hashTokenForLookup creates a SHA256 hash for fast token lookup
func hashTokenForLookup(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// verifyToken verifies a token against its hash
func verifyToken(token, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	return err == nil
}

// CreateToken creates a new SCIM token
func (s *TokenService) CreateToken(ctx context.Context, tenantID uuid.UUID, req *CreateTokenRequest) (*models.SCIMToken, string, error) {
	// Generate token
	plaintextToken, err := generateToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash token (bcrypt for verification)
	tokenHash, err := hashToken(plaintextToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash token: %w", err)
	}

	// Create lookup hash (SHA256 for fast lookup)
	lookupHash := hashTokenForLookup(plaintextToken)

	// Parse expires_at if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, "", fmt.Errorf("invalid expires_at format: %w", err)
		}
		expiresAt = &parsed
	}

	// Get actor from context (if available)
	var createdBy *uuid.UUID
	// TODO: Extract from context if available

	token := &models.SCIMToken{
		ID:         uuid.New(),
		TenantID:   tenantID,
		Name:       req.Name,
		TokenHash:  tokenHash,
		LookupHash: lookupHash,
		Scopes:     req.Scopes,
		ExpiresAt:  expiresAt,
		CreatedBy:  createdBy,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, "", fmt.Errorf("failed to create token: %w", err)
	}

	return token, plaintextToken, nil
}

// GetToken retrieves a SCIM token by ID
func (s *TokenService) GetToken(ctx context.Context, id uuid.UUID) (*models.SCIMToken, error) {
	return s.tokenRepo.GetByID(ctx, id)
}

// ListTokens lists SCIM tokens for a tenant
func (s *TokenService) ListTokens(ctx context.Context, tenantID uuid.UUID) ([]*models.SCIMToken, error) {
	return s.tokenRepo.List(ctx, tenantID)
}

// UpdateToken updates a SCIM token
func (s *TokenService) UpdateToken(ctx context.Context, id uuid.UUID, req *UpdateTokenRequest) (*models.SCIMToken, error) {
	// Get existing token
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		token.Name = *req.Name
	}

	if req.Scopes != nil {
		if len(req.Scopes) == 0 {
			return nil, fmt.Errorf("at least one scope is required")
		}
		token.Scopes = req.Scopes
	}

	if req.ExpiresAt != nil {
		parsed, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format: %w", err)
		}
		token.ExpiresAt = &parsed
	}

	// Update in database
	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}

	return token, nil
}

// DeleteToken deletes a SCIM token
func (s *TokenService) DeleteToken(ctx context.Context, id uuid.UUID) error {
	return s.tokenRepo.Delete(ctx, id)
}

// ValidateToken validates a SCIM token and returns the token if valid
func (s *TokenService) ValidateToken(ctx context.Context, tokenString string) (*models.SCIMToken, error) {
	// Create lookup hash for fast search
	lookupHash := hashTokenForLookup(tokenString)

	// Get token by lookup hash
	token, err := s.tokenRepo.GetByLookupHash(ctx, lookupHash)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Verify token using bcrypt
	if !verifyToken(tokenString, token.TokenHash) {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is expired
	if token.IsExpired() {
		return nil, fmt.Errorf("token expired")
	}

	// Check if token is deleted
	if token.IsDeleted() {
		return nil, fmt.Errorf("token deleted")
	}

	// Update last used timestamp
	_ = s.tokenRepo.UpdateLastUsed(ctx, token.ID)

	return token, nil
}

