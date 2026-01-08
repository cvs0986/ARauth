package claims

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Builder builds JWT claims from user, roles, and permissions
type Builder struct {
	roleRepo        interfaces.RoleRepository
	permissionRepo  interfaces.PermissionRepository
	systemRoleRepo  interfaces.SystemRoleRepository // NEW: For SYSTEM users
}

// NewBuilder creates a new claims builder
func NewBuilder(roleRepo interfaces.RoleRepository, permissionRepo interfaces.PermissionRepository, systemRoleRepo interfaces.SystemRoleRepository) *Builder {
	return &Builder{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		systemRoleRepo: systemRoleRepo,
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
	PrincipalType    string   `json:"principal_type"` // NEW: SYSTEM, TENANT, SERVICE
	TenantID         string   `json:"tenant_id,omitempty"` // Optional for SYSTEM users
	Email            string   `json:"email,omitempty"`
	Username         string   `json:"username,omitempty"`
	Roles            []string `json:"roles,omitempty"` // Tenant roles
	Permissions      []string `json:"permissions,omitempty"` // Tenant permissions
	SystemRoles      []string `json:"system_roles,omitempty"` // NEW: System roles
	SystemPermissions []string `json:"system_permissions,omitempty"` // NEW: System permissions
	Scope            string   `json:"scope,omitempty"` // Space-separated scopes
}

// BuildClaims builds claims for a user
func (b *Builder) BuildClaims(ctx context.Context, user *models.User) (*Claims, error) {
	claims := &Claims{
		Subject:       user.ID.String(),
		PrincipalType: string(user.PrincipalType),
		Email:         user.Email,
		Username:      user.Username,
		Roles:         []string{},
		Permissions:   []string{},
		SystemRoles:  []string{},
		SystemPermissions: []string{},
	}

	// Handle tenant_id (nullable for SYSTEM users)
	if user.TenantID != nil {
		claims.TenantID = user.TenantID.String()
	}

	// For SYSTEM users: get system roles and permissions
	if user.PrincipalType == models.PrincipalTypeSystem {
		systemRoles, err := b.systemRoleRepo.GetUserSystemRoles(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get system roles: %w", err)
		}

		systemRoleNames := make([]string, 0, len(systemRoles))
		systemPermissionMap := make(map[string]bool)

		for _, role := range systemRoles {
			systemRoleNames = append(systemRoleNames, role.Name)

			// Get permissions for this system role
			permissions, err := b.systemRoleRepo.GetRolePermissions(ctx, role.ID)
			if err != nil {
				continue
			}

			for _, perm := range permissions {
				permissionKey := perm.Resource + ":" + perm.Action
				systemPermissionMap[permissionKey] = true
			}
		}

		claims.SystemRoles = systemRoleNames

		// Convert system permission map to slice
		systemPermissions := make([]string, 0, len(systemPermissionMap))
		for perm := range systemPermissionMap {
			systemPermissions = append(systemPermissions, perm)
		}
		claims.SystemPermissions = systemPermissions

		// Build scope for SYSTEM users
		scopeParts := make([]string, 0)
		scopeParts = append(scopeParts, "system:*")
		for _, role := range systemRoleNames {
			scopeParts = append(scopeParts, "system_role:"+role)
		}
		claims.Scope = joinStrings(scopeParts, " ")

		return claims, nil
	}

	// For TENANT users: get tenant roles and permissions
	// Get user roles
	roles, err := b.roleRepo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant roles: %w", err)
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

	// Build scope string for tenant users
	scopeParts := make([]string, 0)
	if user.TenantID != nil {
		scopeParts = append(scopeParts, "tenant:"+user.TenantID.String())
	}
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

