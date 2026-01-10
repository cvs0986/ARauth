package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/config"
	"golang.org/x/crypto/bcrypt"
)

// Service provides token generation and validation
type Service struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	secret        []byte // Fallback for HS256
	issuer        string
	lifetimeResolver *LifetimeResolver
}

// NewService creates a new token service
func NewService(cfg *config.SecurityConfig, lifetimeResolver *LifetimeResolver) (*Service, error) {
	service := &Service{
		issuer:          cfg.JWT.Issuer,
		lifetimeResolver: lifetimeResolver,
	}

	// Try to load RSA key pair
	if cfg.JWT.SigningKeyPath != "" {
		privateKey, publicKey, err := loadRSAKeyPair(cfg.JWT.SigningKeyPath)
		if err == nil {
			service.privateKey = privateKey
			service.publicKey = publicKey
			return service, nil
		}
		// If loading fails, fall back to HS256
	}

	// Fallback to HS256 with secret
	if cfg.JWT.Secret != "" {
		service.secret = []byte(cfg.JWT.Secret)
		return service, nil
	}

	// Generate a temporary RSA key pair for development
	// In production, this should be configured properly
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	service.privateKey = privateKey
	service.publicKey = &privateKey.PublicKey

	return service, nil
}

// GenerateAccessToken generates a JWT access token
func (s *Service) GenerateAccessToken(claimsObj *claims.Claims, expiresIn time.Duration) (string, error) {
	now := time.Now()

	// Build JWT claims
	tokenClaims := jwt.MapClaims{
		"sub":                claimsObj.Subject,
		"principal_type":     claimsObj.PrincipalType, // NEW: Include principal_type
		"tenant_id":          claimsObj.TenantID,
		"email":              claimsObj.Email,
		"username":           claimsObj.Username,
		"roles":              claimsObj.Roles,
		"permissions":        claimsObj.Permissions,
		"system_roles":       claimsObj.SystemRoles,       // NEW: Include system_roles
		"system_permissions": claimsObj.SystemPermissions, // NEW: Include system_permissions
		"scope":              claimsObj.Scope,
		"iss":                s.issuer,
		"iat":                now.Unix(),
		"exp":                now.Add(expiresIn).Unix(),
		"jti":                uuid.New().String(),
	}

	// Add impersonation claims if present
	if claimsObj.ImpersonatedBy != "" {
		tokenClaims["impersonated_by"] = claimsObj.ImpersonatedBy
	}
	if claimsObj.ImpersonationSessionID != "" {
		tokenClaims["impersonation_session_id"] = claimsObj.ImpersonationSessionID
	}

	// Create token
	var token *jwt.Token
	if s.privateKey != nil {
		// Use RS256
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, tokenClaims)
		return token.SignedString(s.privateKey)
	} else if len(s.secret) > 0 {
		// Fallback to HS256
		token = jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
		return token.SignedString(s.secret)
	}

	return "", fmt.Errorf("no signing key available")
}

// GenerateRefreshToken generates an opaque refresh token (UUID)
func (s *Service) GenerateRefreshToken() (string, error) {
	return uuid.New().String(), nil
}

// HashRefreshToken hashes a refresh token for storage
func (s *Service) HashRefreshToken(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash refresh token: %w", err)
	}
	return string(hash), nil
}

// VerifyRefreshToken verifies a refresh token against its hash
func (s *Service) VerifyRefreshToken(token, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	return err == nil
}

// ValidateAccessToken validates and parses an access token
func (s *Service) ValidateAccessToken(tokenString string) (*claims.Claims, error) {
	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); ok {
			if s.publicKey == nil {
				return nil, fmt.Errorf("public key not available")
			}
			return s.publicKey, nil
		}
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
			if len(s.secret) == 0 {
				return nil, fmt.Errorf("secret not available")
			}
			return s.secret, nil
		}
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Build claims object
	claimsObj := &claims.Claims{
		Subject:       getStringClaim(claimsMap, "sub"),
		PrincipalType: getStringClaim(claimsMap, "principal_type"), // NEW: Extract principal_type
		TenantID:      getStringClaim(claimsMap, "tenant_id"),
		Email:         getStringClaim(claimsMap, "email"),
		Username:      getStringClaim(claimsMap, "username"),
		Issuer:        getStringClaim(claimsMap, "iss"),
		Audience:      getStringClaim(claimsMap, "aud"),
	}

	// Extract roles
	if roles, ok := claimsMap["roles"].([]interface{}); ok {
		claimsObj.Roles = make([]string, len(roles))
		for i, role := range roles {
			if r, ok := role.(string); ok {
				claimsObj.Roles[i] = r
			}
		}
	}

	// Extract permissions
	if perms, ok := claimsMap["permissions"].([]interface{}); ok {
		claimsObj.Permissions = make([]string, len(perms))
		for i, perm := range perms {
			if p, ok := perm.(string); ok {
				claimsObj.Permissions[i] = p
			}
		}
	}

	// Extract system_roles
	if systemRoles, ok := claimsMap["system_roles"].([]interface{}); ok {
		claimsObj.SystemRoles = make([]string, len(systemRoles))
		for i, role := range systemRoles {
			if r, ok := role.(string); ok {
				claimsObj.SystemRoles[i] = r
			}
		}
	}

	// Extract system_permissions
	if systemPerms, ok := claimsMap["system_permissions"].([]interface{}); ok {
		claimsObj.SystemPermissions = make([]string, len(systemPerms))
		for i, perm := range systemPerms {
			if p, ok := perm.(string); ok {
				claimsObj.SystemPermissions[i] = p
			}
		}
	}

	// Extract scope
	claimsObj.Scope = getStringClaim(claimsMap, "scope")

	// Extract timestamps
	if exp, ok := claimsMap["exp"].(float64); ok {
		claimsObj.ExpiresAt = int64(exp)
	}
	if iat, ok := claimsMap["iat"].(float64); ok {
		claimsObj.IssuedAt = int64(iat)
	}

	return claimsObj, nil
}

// getStringClaim safely extracts a string claim
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// loadRSAKeyPair loads RSA key pair from file
func loadRSAKeyPair(keyPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// Load private key
	privateKeyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read private key: %w", err)
	}

	block, _ := pem.Decode(privateKeyData)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode PEM block")
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		parsedKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, fmt.Errorf("key is not RSA private key")
		}
	}

	publicKey := &parsedKey.PublicKey

	return parsedKey, publicKey, nil
}

// GetPublicKey returns the public key for JWKS endpoint
func (s *Service) GetPublicKey() interface{} {
	if s.publicKey != nil {
		return s.publicKey
	}
	return nil
}

