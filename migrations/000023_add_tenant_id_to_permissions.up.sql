-- Migration: Add tenant_id to permissions table
-- This makes permissions tenant-scoped, allowing each tenant to have their own permissions

-- Add tenant_id column (nullable for backward compatibility, but should be NOT NULL for new permissions)
ALTER TABLE permissions 
ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create index for tenant_id
CREATE INDEX idx_permissions_tenant_id ON permissions(tenant_id);

-- Update unique constraint to include tenant_id
-- First, drop the existing unique constraint on (resource, action)
ALTER TABLE permissions DROP CONSTRAINT IF EXISTS permissions_resource_action_key;

-- Add new unique constraint on (tenant_id, resource, action)
-- Note: tenant_id can be NULL for backward compatibility with existing global permissions
CREATE UNIQUE INDEX idx_permissions_tenant_resource_action 
ON permissions(tenant_id, resource, action) 
WHERE tenant_id IS NOT NULL;

-- For NULL tenant_id (global permissions), keep unique on (resource, action)
CREATE UNIQUE INDEX idx_permissions_global_resource_action 
ON permissions(resource, action) 
WHERE tenant_id IS NULL;

-- Add updated_at and deleted_at columns for consistency with other tables
ALTER TABLE permissions 
ADD COLUMN updated_at TIMESTAMP DEFAULT NOW(),
ADD COLUMN deleted_at TIMESTAMP;

-- Create index for deleted_at
CREATE INDEX idx_permissions_deleted_at ON permissions(deleted_at) WHERE deleted_at IS NULL;

