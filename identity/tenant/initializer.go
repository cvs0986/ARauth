package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/arauth-identity/iam/identity/models"
	"github.com/arauth-identity/iam/storage/interfaces"
)

// Initializer handles initialization of predefined roles and permissions for new tenants
type Initializer struct {
	roleRepo       interfaces.RoleRepository
	permissionRepo interfaces.PermissionRepository
}

// NewInitializer creates a new tenant initializer
func NewInitializer(roleRepo interfaces.RoleRepository, permissionRepo interfaces.PermissionRepository) *Initializer {
	return &Initializer{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// InitializationResult contains the results of tenant initialization
type InitializationResult struct {
	TenantOwnerRoleID  uuid.UUID
	TenantAdminRoleID  uuid.UUID
	TenantAuditorRoleID uuid.UUID
	PermissionsCreated int
}

// InitializeTenant creates predefined roles and permissions for a new tenant
func (i *Initializer) InitializeTenant(ctx context.Context, tenantID uuid.UUID) (*InitializationResult, error) {
	// 1. Create predefined permissions
	permissions, err := i.createPredefinedPermissions(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create permissions: %w", err)
	}

	// 2. Create predefined roles
	roles, err := i.createPredefinedRoles(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create roles: %w", err)
	}

	// 3. Assign permissions to roles
	if err := i.assignPermissionsToRoles(ctx, roles, permissions); err != nil {
		return nil, fmt.Errorf("failed to assign permissions: %w", err)
	}

	return &InitializationResult{
		TenantOwnerRoleID:  roles["tenant_owner"].ID,
		TenantAdminRoleID:  roles["tenant_admin"].ID,
		TenantAuditorRoleID: roles["tenant_auditor"].ID,
		PermissionsCreated: len(permissions),
	}, nil
}

// createPredefinedPermissions creates all predefined permissions for a tenant
func (i *Initializer) createPredefinedPermissions(ctx context.Context, tenantID uuid.UUID) (map[string]*models.Permission, error) {
	permissions := make(map[string]*models.Permission)

	// Define all predefined permissions
	// Using tenant.* namespace to avoid collision with system permissions
	permissionDefs := []struct {
		key         string
		resource    string
		action      string
		description string
	}{
		// User Management
		{"tenant.users.create", "tenant.users", "create", "Create new users"},
		{"tenant.users.read", "tenant.users", "read", "View users"},
		{"tenant.users.update", "tenant.users", "update", "Update users"},
		{"tenant.users.delete", "tenant.users", "delete", "Delete users"},
		{"tenant.users.manage", "tenant.users", "manage", "Full user management"},

		// Role Management
		{"tenant.roles.create", "tenant.roles", "create", "Create roles"},
		{"tenant.roles.read", "tenant.roles", "read", "View roles"},
		{"tenant.roles.update", "tenant.roles", "update", "Update roles"},
		{"tenant.roles.delete", "tenant.roles", "delete", "Delete roles"},
		{"tenant.roles.manage", "tenant.roles", "manage", "Full role management"},

		// Permission Management (only for tenant_owner by default)
		{"tenant.permissions.create", "tenant.permissions", "create", "Create permissions"},
		{"tenant.permissions.read", "tenant.permissions", "read", "View permissions"},
		{"tenant.permissions.update", "tenant.permissions", "update", "Update permissions"},
		{"tenant.permissions.delete", "tenant.permissions", "delete", "Delete permissions"},
		{"tenant.permissions.manage", "tenant.permissions", "manage", "Full permission management"},

		// Tenant Settings
		{"tenant.settings.read", "tenant.settings", "read", "View tenant settings"},
		{"tenant.settings.update", "tenant.settings", "update", "Update tenant settings"},

		// Audit & Logs
		{"tenant.audit.read", "tenant.audit", "read", "View audit logs"},

		// Admin Access
		{"tenant.admin.access", "tenant.admin", "access", "Access admin dashboard"},
	}

	for _, def := range permissionDefs {
		// Check if permission already exists (idempotent) by listing with filters
		filters := &interfaces.PermissionFilters{
			Resource: &def.resource,
			Action:   &def.action,
		}
		existingList, err := i.permissionRepo.List(ctx, tenantID, filters)
		if err == nil && len(existingList) > 0 {
			// Check if any existing permission matches tenant_id
			for _, existing := range existingList {
				if existing.TenantID == tenantID {
					permissions[def.key] = existing
					break
				}
			}
			if permissions[def.key] != nil {
				continue
			}
		}

		desc := def.description
		permission := &models.Permission{
			TenantID:    tenantID,
			Resource:    def.resource,
			Action:      def.action,
			Description: &desc,
		}

		if err := i.permissionRepo.Create(ctx, permission); err != nil {
			return nil, fmt.Errorf("failed to create permission %s: %w", def.key, err)
		}

		permissions[def.key] = permission
	}

	return permissions, nil
}

// createPredefinedRoles creates predefined roles for a tenant
func (i *Initializer) createPredefinedRoles(ctx context.Context, tenantID uuid.UUID) (map[string]*models.Role, error) {
	roles := make(map[string]*models.Role)

	roleDefs := []struct {
		name        string
		description string
	}{
		{"tenant_owner", "Full tenant ownership and control. Can manage everything in the tenant."},
		{"tenant_admin", "Tenant administration with most admin features. Cannot delete tenant or modify tenant owner."},
		{"tenant_auditor", "Read-only access for auditing and compliance. Can view all data but cannot modify."},
	}

	for _, def := range roleDefs {
		// Check if role already exists (idempotent)
		existing, err := i.roleRepo.GetByName(ctx, tenantID, def.name)
		if err == nil && existing != nil {
			roles[def.name] = existing
			continue
		}

		desc := def.description
		role := &models.Role{
			TenantID:    tenantID,
			Name:        def.name,
			Description: &desc,
			IsSystem:    true, // Mark as system role (non-deletable, non-modifiable)
		}

		if err := i.roleRepo.Create(ctx, role); err != nil {
			return nil, fmt.Errorf("failed to create role %s: %w", def.name, err)
		}

		roles[def.name] = role
	}

	return roles, nil
}

// assignPermissionsToRoles assigns permissions to predefined roles
func (i *Initializer) assignPermissionsToRoles(ctx context.Context, roles map[string]*models.Role, permissions map[string]*models.Permission) error {
	// Tenant Owner gets all permissions
	tenantOwner := roles["tenant_owner"]
	for _, perm := range permissions {
		if err := i.permissionRepo.AssignPermissionToRole(ctx, tenantOwner.ID, perm.ID); err != nil {
			// Check if it's a duplicate key error (idempotent)
			if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "duplicate") {
				return fmt.Errorf("failed to assign permission %s to tenant_owner: %w", perm.Resource+":"+perm.Action, err)
			}
		}
	}

	// Tenant Admin gets most permissions (except tenant deletion and permission management)
	// Note: permissions:* removed by default - many enterprises prefer security team controls permissions
	// Tenants can add it later if needed
	tenantAdmin := roles["tenant_admin"]
	adminPermissions := []string{
		"tenant.users.create", "tenant.users.read", "tenant.users.update", "tenant.users.delete", "tenant.users.manage",
		"tenant.roles.create", "tenant.roles.read", "tenant.roles.update", "tenant.roles.delete", "tenant.roles.manage",
		// Note: permissions:* removed - tenant_admin cannot manage permissions by default
		"tenant.permissions.read", // Read-only access to permissions
		"tenant.settings.read", "tenant.settings.update",
		"tenant.audit.read",
		"tenant.admin.access",
	}
	for _, permKey := range adminPermissions {
		if perm, exists := permissions[permKey]; exists {
			if err := i.permissionRepo.AssignPermissionToRole(ctx, tenantAdmin.ID, perm.ID); err != nil {
				if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "duplicate") {
					return fmt.Errorf("failed to assign permission %s to tenant_admin: %w", permKey, err)
				}
			}
		}
	}

	// Tenant Auditor gets read-only permissions
	tenantAuditor := roles["tenant_auditor"]
	auditorPermissions := []string{
		"tenant.users.read",
		"tenant.roles.read",
		"tenant.permissions.read",
		"tenant.audit.read",
		"tenant.admin.access",
	}
	for _, permKey := range auditorPermissions {
		if perm, exists := permissions[permKey]; exists {
			if err := i.permissionRepo.AssignPermissionToRole(ctx, tenantAuditor.ID, perm.ID); err != nil {
				if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "duplicate") {
					return fmt.Errorf("failed to assign permission %s to tenant_auditor: %w", permKey, err)
				}
			}
		}
	}

	return nil
}

// AttachAllPermissionsToTenantOwner ensures tenant_owner has all current tenant permissions
// This should be called when new permissions are created to maintain the invariant:
// "tenant_owner always has all tenant permissions"
func (i *Initializer) AttachAllPermissionsToTenantOwner(ctx context.Context, tenantID uuid.UUID) error {
	// Get tenant_owner role
	tenantOwnerRole, err := i.roleRepo.GetByName(ctx, tenantID, "tenant_owner")
	if err != nil {
		return fmt.Errorf("tenant_owner role not found: %w", err)
	}

	// Get all permissions for this tenant
	filters := &interfaces.PermissionFilters{
		Page:     1,
		PageSize: 1000, // Get all permissions
	}
	allPermissions, err := i.permissionRepo.List(ctx, tenantID, filters)
	if err != nil {
		return fmt.Errorf("failed to list tenant permissions: %w", err)
	}

	// Assign all permissions to tenant_owner
	for _, perm := range allPermissions {
		if err := i.permissionRepo.AssignPermissionToRole(ctx, tenantOwnerRole.ID, perm.ID); err != nil {
			// Ignore duplicate assignment errors (idempotent)
			if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "duplicate") {
				return fmt.Errorf("failed to assign permission %s to tenant_owner: %w", perm.Resource+":"+perm.Action, err)
			}
		}
	}

	return nil
}

