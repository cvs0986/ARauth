package claims

import (
	"context"

	"github.com/google/uuid"
	"github.com/nuage-identity/iam/identity/models"
	"github.com/nuage-identity/iam/storage/interfaces"
)

// Builder builds JWT claims from user, roles, and permissions
type Builder struct {
	roleRepo       interfaces.RoleRepository
	permissionRepo interfaces.PermissionRepository
}

// NewBuilder creates a new claims builder
func NewBuilder(roleRepo interfaces.RoleRepository, permissionRepo interfaces.PermissionRepository) *Builder {
	return &Builder{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// Claims represents JWT claims
type Claims struct {
	// Standard claims
	Subject string `json:"sub"` // User ID
	Issuer  string `json:"iss,omitempty"`
	Audience string `json:"aud,omitempty"`
	ExpiresAt int64 `json:"exp,omitempty"`
	IssuedAt  int64 `json:"iat,omitempty"`
	NotBefore int64 `json:"nbf,omitempty"`

	// Custom claims
	TenantID   string   `json:"tenant_id"`
	Email      string   `json:"email,omitempty"`
	Username   string   `json:"username,omitempty"`
	Roles      []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Scope      string   `json:"scope,omitempty"` // Space-separated scopes
}

// BuildClaims builds claims for a user
func (b *Builder) BuildClaims(ctx context.Context, user *models.User) (*Claims, error) {
	claims := &Claims{
		Subject:  user.ID.String(),
		TenantID: user.TenantID.String(),
		Email:    user.Email,
		Username: user.Username,
		Roles:    []string{},
		Permissions: []string{},
	}

	// Get user roles
	roles, err := b.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	// Extract role names
	roleNames := make([]string, 0, len(roles))
	permissionMap := make(map[string]bool) // Use map to avoid duplicates

	for _, role := range roles {
		roleNames = append(roleNames, role.Name)

		// Get permissions for this role
		permissions, err := b.permissionRepo.GetRolePermissions(ctx, role.ID)
		if err != nil {
			continue // Continue even if we can't get permissions for one role
		}

		// Add permissions to map
		for _, perm := range permissions {
			permissionKey := perm.Resource + ":" + perm.Action
			permissionMap[permissionKey] = true
		}
	}

	claims.Roles = roleNames

	// Convert permission map to slice
	permissions := make([]string, 0, len(permissionMap))
	for perm := range permissionMap {
		permissions = append(permissions, perm)
	}
	claims.Permissions = permissions

	// Build scope string (space-separated)
	scopeParts := make([]string, 0)
	for _, role := range roleNames {
		scopeParts = append(scopeParts, "role:"+role)
	}
	for perm := range permissionMap {
		scopeParts = append(scopeParts, "perm:"+perm)
	}
	claims.Scope = joinStrings(scopeParts, " ")

	return claims, nil
}

// BuildClaimsForUserID builds claims for a user by ID
func (b *Builder) BuildClaimsForUserID(ctx context.Context, userID uuid.UUID) (*Claims, error) {
	// This method would need user repository to get user first
	// For now, we'll require the user object to be passed
	return nil, nil
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

