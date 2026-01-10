package introspection

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/arauth-identity/iam/auth/claims"
	"github.com/arauth-identity/iam/auth/token"
)

// Service provides token introspection functionality (RFC 7662)
type Service struct {
	tokenService token.ServiceInterface
	claimsParser *claims.Parser
}

// NewService creates a new token introspection service
func NewService(tokenService token.ServiceInterface, claimsParser *claims.Parser) ServiceInterface {
	return &Service{
		tokenService: tokenService,
		claimsParser: claimsParser,
	}
}

// IntrospectToken introspects a token and returns its metadata
func (s *Service) IntrospectToken(ctx context.Context, tokenString string, tokenTypeHint string) (*TokenInfo, error) {
	// Parse and validate the token
	token, err := s.tokenService.ValidateToken(ctx, tokenString)
	if err != nil {
		// Token is invalid or expired - return inactive token
		return &TokenInfo{
			Active: false,
		}, nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &TokenInfo{
			Active: false,
		}, nil
	}

	// Check if token is expired
	exp, ok := claims["exp"].(float64)
	if ok {
		expTime := time.Unix(int64(exp), 0)
		if time.Now().After(expTime) {
			return &TokenInfo{
				Active: false,
			}, nil
		}
	}

	// Build token info
	info := &TokenInfo{
		Active: true,
	}

	// Extract standard claims
	if sub, ok := claims["sub"].(string); ok {
		info.Subject = sub
	}
	if iss, ok := claims["iss"].(string); ok {
		info.Issuer = iss
	}
	if aud, ok := claims["aud"].(string); ok {
		info.Audience = aud
	}
	if jti, ok := claims["jti"].(string); ok {
		info.JTI = jti
	}
	if username, ok := claims["username"].(string); ok {
		info.Username = username
	}

	// Extract timestamps
	if exp, ok := claims["exp"].(float64); ok {
		info.ExpiresAt = int64(exp)
	}
	if iat, ok := claims["iat"].(float64); ok {
		info.IssuedAt = int64(iat)
	}
	if nbf, ok := claims["nbf"].(float64); ok {
		info.NotBefore = int64(nbf)
	}

	// Extract scope
	if scope, ok := claims["scope"].(string); ok {
		info.Scope = scope
	}

	// Extract client ID (if available)
	if clientID, ok := claims["client_id"].(string); ok {
		info.ClientID = clientID
	}

	// Extract ARauth-specific claims
	if tenantID, ok := claims["tenant_id"].(string); ok {
		info.TenantID = tenantID
	}
	if principalType, ok := claims["principal_type"].(string); ok {
		info.PrincipalType = principalType
	}

	// Extract roles
	if roles, ok := claims["roles"].([]interface{}); ok {
		info.Roles = make([]string, 0, len(roles))
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				info.Roles = append(info.Roles, roleStr)
			}
		}
	}

	// Extract permissions
	if perms, ok := claims["permissions"].([]interface{}); ok {
		info.Permissions = make([]string, 0, len(perms))
		for _, perm := range perms {
			if permStr, ok := perm.(string); ok {
				info.Permissions = append(info.Permissions, permStr)
			}
		}
	}

	// Extract system roles (SYSTEM users)
	if systemRoles, ok := claims["system_roles"].([]interface{}); ok {
		info.SystemRoles = make([]string, 0, len(systemRoles))
		for _, role := range systemRoles {
			if roleStr, ok := role.(string); ok {
				info.SystemRoles = append(info.SystemRoles, roleStr)
			}
		}
	}

	// Extract system permissions (SYSTEM users)
	if systemPerms, ok := claims["system_permissions"].([]interface{}); ok {
		info.SystemPerms = make([]string, 0, len(systemPerms))
		for _, perm := range systemPerms {
			if permStr, ok := perm.(string); ok {
				info.SystemPerms = append(info.SystemPerms, permStr)
			}
		}
	}

	return info, nil
}

