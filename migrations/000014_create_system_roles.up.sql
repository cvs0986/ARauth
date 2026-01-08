-- Migration: Create system roles and permissions tables
-- System roles are separate from tenant roles and apply to SYSTEM principal type users

-- System roles table
CREATE TABLE system_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- System permissions table
CREATE TABLE system_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- System role permissions junction table
CREATE TABLE system_role_permissions (
    role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES system_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- User system roles junction table
CREATE TABLE user_system_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES system_roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Create indexes
CREATE INDEX idx_system_roles_name ON system_roles(name);
CREATE INDEX idx_system_permissions_resource_action ON system_permissions(resource, action);
CREATE INDEX idx_system_role_permissions_role_id ON system_role_permissions(role_id);
CREATE INDEX idx_system_role_permissions_permission_id ON system_role_permissions(permission_id);
CREATE INDEX idx_user_system_roles_user_id ON user_system_roles(user_id);
CREATE INDEX idx_user_system_roles_role_id ON user_system_roles(role_id);

-- Insert default system roles (using fixed UUIDs for consistency)
INSERT INTO system_roles (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'system_owner', 'Full system ownership and control. Can manage all aspects of the platform.'),
    ('00000000-0000-0000-0000-000000000002', 'system_admin', 'System administration with tenant management capabilities. Cannot delete system or modify system owner.'),
    ('00000000-0000-0000-0000-000000000003', 'system_auditor', 'Read-only system access for auditing and compliance. Can view all system data but cannot modify.');

-- Insert default system permissions
INSERT INTO system_permissions (resource, action, description) VALUES
    ('tenant', 'create', 'Create new tenants'),
    ('tenant', 'read', 'View all tenants and their details'),
    ('tenant', 'update', 'Update any tenant configuration'),
    ('tenant', 'delete', 'Delete any tenant'),
    ('tenant', 'suspend', 'Suspend tenant access'),
    ('tenant', 'resume', 'Resume suspended tenant'),
    ('tenant', 'configure', 'Configure tenant-specific settings'),
    ('system', 'settings', 'Manage system-wide settings'),
    ('system', 'policy', 'Manage global security policies'),
    ('system', 'audit', 'View system audit logs'),
    ('system', 'users', 'Manage system users'),
    ('billing', 'manage', 'Manage billing and subscriptions'),
    ('billing', 'read', 'View billing information');

-- Assign all permissions to system_owner
INSERT INTO system_role_permissions (role_id, permission_id)
SELECT 
    '00000000-0000-0000-0000-000000000001'::uuid,
    id
FROM system_permissions;

-- Assign permissions to system_admin (tenant management + system settings, but not delete system or modify system owner)
INSERT INTO system_role_permissions (role_id, permission_id)
SELECT 
    '00000000-0000-0000-0000-000000000002'::uuid,
    id
FROM system_permissions
WHERE (resource = 'tenant' AND action != 'delete') OR 
      (resource = 'system' AND action IN ('settings', 'policy', 'audit')) OR
      resource = 'billing';

-- Assign permissions to system_auditor (read-only)
INSERT INTO system_role_permissions (role_id, permission_id)
SELECT 
    '00000000-0000-0000-0000-000000000003'::uuid,
    id
FROM system_permissions
WHERE action = 'read' OR action = 'audit';

