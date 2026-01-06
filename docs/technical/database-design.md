# Database Design

This document describes the database schema, relationships, and design decisions for Nuage Identity.

## 🎯 Design Principles

1. **Normalization**: 3NF where possible
2. **Performance**: Indexes on frequently queried columns
3. **Scalability**: Support for horizontal scaling
4. **Multi-tenancy**: Tenant isolation built-in
5. **Audit**: Created/updated timestamps on all tables

## 📊 Entity Relationship Diagram

```
┌─────────────┐
│   Tenants   │
└──────┬──────┘
       │
       │ 1:N
       │
┌──────▼──────┐      ┌─────────────┐
│    Users    │──────│ Credentials │
└──────┬──────┘ 1:1  └─────────────┘
       │
       │ N:M
       │
┌──────▼──────┐      ┌─────────────┐
│ User Roles  │──────│    Roles    │
└─────────────┘      └──────┬──────┘
                            │
                            │ N:M
                            │
                   ┌────────▼────────┐
                   │ Role Permissions│
                   └────────┬───────┘
                            │
                            │ N:1
                            │
                   ┌────────▼────────┐
                   │  Permissions   │
                   └────────────────┘
```

## 📋 Schema Definition

### Tenants Table

```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_tenants_domain ON tenants(domain);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NULL;
```

**Fields**:
- `id`: Primary key (UUID)
- `name`: Tenant name
- `domain`: Unique domain identifier
- `status`: active, suspended, deleted
- `metadata`: JSON for custom fields
- `created_at`, `updated_at`: Audit timestamps
- `deleted_at`: Soft delete

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret_encrypted TEXT,
    last_login_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, username),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_username_tenant ON users(tenant_id, username);
CREATE INDEX idx_users_email_tenant ON users(tenant_id, email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
```

**Fields**:
- `id`: Primary key (UUID)
- `tenant_id`: Foreign key to tenants
- `username`: Unique within tenant
- `email`: Unique within tenant
- `status`: active, suspended, deleted
- `mfa_enabled`: MFA status
- `mfa_secret_encrypted`: Encrypted TOTP secret
- `last_login_at`: Last login timestamp
- `metadata`: JSON for custom fields

### Credentials Table

```sql
CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    password_changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    password_expires_at TIMESTAMP,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

CREATE INDEX idx_credentials_user_id ON credentials(user_id);
```

**Fields**:
- `id`: Primary key (UUID)
- `user_id`: Foreign key to users (1:1)
- `password_hash`: Argon2id hash
- `password_changed_at`: Password change timestamp
- `password_expires_at`: Password expiration (optional)
- `failed_login_attempts`: Failed login counter
- `locked_until`: Account lock timestamp

### Roles Table

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_name_tenant ON roles(tenant_id, name);
CREATE INDEX idx_roles_deleted_at ON roles(deleted_at) WHERE deleted_at IS NULL;
```

**Fields**:
- `id`: Primary key (UUID)
- `tenant_id`: Foreign key to tenants
- `name`: Unique within tenant
- `description`: Role description
- `is_system`: System role (cannot be deleted)

### Permissions Table

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(resource, action)
);

CREATE INDEX idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX idx_permissions_name ON permissions(name);
```

**Fields**:
- `id`: Primary key (UUID)
- `name`: Unique permission name (e.g., "user.read")
- `description`: Permission description
- `resource`: Resource type (e.g., "user", "tenant")
- `action`: Action type (e.g., "read", "write", "delete")

### User Roles Table (Junction)

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

**Fields**:
- `id`: Primary key (UUID)
- `user_id`: Foreign key to users
- `role_id`: Foreign key to roles
- `assigned_at`: Assignment timestamp
- `assigned_by`: User who assigned the role

### Role Permissions Table (Junction)

```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
```

**Fields**:
- `id`: Primary key (UUID)
- `role_id`: Foreign key to roles
- `permission_id`: Foreign key to permissions

### MFA Recovery Codes Table

```sql
CREATE TABLE mfa_recovery_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash TEXT NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mfa_recovery_codes_user_id ON mfa_recovery_codes(user_id);
CREATE INDEX idx_mfa_recovery_codes_used ON mfa_recovery_codes(user_id, used_at) WHERE used_at IS NULL;
```

**Fields**:
- `id`: Primary key (UUID)
- `user_id`: Foreign key to users
- `code_hash`: Hashed recovery code
- `used_at`: Usage timestamp (NULL if unused)

### Audit Log Table

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    actor_id UUID REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(255) NOT NULL,
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

**Fields**:
- `id`: Primary key (UUID)
- `tenant_id`: Tenant context
- `actor_id`: User who performed action
- `action`: Action type (e.g., "user.created")
- `resource_type`: Resource type
- `resource_id`: Resource ID
- `ip_address`: Client IP
- `user_agent`: Client user agent
- `metadata`: Additional data (JSON)

## 🔑 Indexes Strategy

### Primary Indexes

All tables have UUID primary keys with default generation.

### Foreign Key Indexes

All foreign keys are indexed for join performance.

### Composite Indexes

**Critical Composite Indexes**:
- `users(tenant_id, username)`: User lookup
- `users(tenant_id, email)`: Email lookup
- `roles(tenant_id, name)`: Role lookup
- `user_roles(user_id, role_id)`: User role queries

### Partial Indexes

**Soft Delete Support**:
- `WHERE deleted_at IS NULL` on all soft-deletable tables

## 🔄 Migration Strategy

### Migration Tool

**Tool**: `golang-migrate/migrate`

**Structure**:
```
migrations/
├── 000001_create_tenants.up.sql
├── 000001_create_tenants.down.sql
├── 000002_create_users.up.sql
├── 000002_create_users.down.sql
...
```

### Migration Best Practices

1. **Idempotent**: Migrations should be safe to run multiple times
2. **Reversible**: Always provide down migrations
3. **Tested**: Test migrations in development first
4. **Backed Up**: Backup before production migrations

## 🔐 Data Security

### Encryption

**Encrypted Fields**:
- `credentials.password_hash`: Already hashed (Argon2id)
- `users.mfa_secret_encrypted`: AES-256-GCM encrypted
- `mfa_recovery_codes.code_hash`: Hashed (bcrypt)

### Access Control

**Database Level**:
- Separate database user for application
- Minimum required permissions
- No direct table access from application (use views if needed)

**Application Level**:
- Tenant isolation in queries
- Row-level security (PostgreSQL RLS) - optional

## 📊 Performance Optimization

### Query Optimization

**Common Queries Optimized**:
1. User lookup by username/tenant
2. User roles and permissions
3. Tenant-scoped queries

**Example Optimized Query**:
```sql
-- Get user with roles and permissions
SELECT 
    u.id, u.username, u.email,
    r.id as role_id, r.name as role_name,
    p.name as permission_name
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
WHERE u.tenant_id = $1 AND u.username = $2 AND u.deleted_at IS NULL;
```

### Connection Pooling

**Configuration**:
- Max connections: 25 per instance
- Idle connections: 5
- Connection lifetime: 5 minutes

## 🔄 Multi-Database Support

### Abstraction Layer

**Repository Interface**:
```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByUsername(ctx context.Context, username string, tenantID string) (*User, error)
}
```

### Database Adapters

**Supported**:
- PostgreSQL (primary)
- MySQL (future)
- MSSQL (future)
- MongoDB (future)

**Adapter Implementation**:
- Each database has its own implementation
- Same interface, different SQL/driver

## 📚 Related Documentation

- [Architecture Overview](../architecture/overview.md)
- [Technical Stack](./tech-stack.md)
- [Security](./security.md)

