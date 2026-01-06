# Component Architecture

This document provides detailed documentation of each component in the Nuage Identity system.

## 📦 Component Overview

```
iam/
├── cmd/
│   └── server/              # Application entry point
├── api/                     # HTTP API layer
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   └── routes/             # Route definitions
├── auth/                   # Authentication service
│   ├── login/             # Login logic
│   ├── mfa/               # MFA implementation
│   ├── refresh/           # Token refresh
│   └── hydra/             # Hydra integration
├── identity/               # Identity management
│   ├── user/              # User management
│   ├── tenant/            # Tenant management
│   ├── group/             # Group management
│   └── credential/        # Credential management
├── policy/                 # Authorization
│   ├── rbac/              # Role-based access control
│   ├── abac/              # Attribute-based access control
│   ├── claims/            # Claims builder
│   └── permissions/       # Permission management
├── storage/                # Data access layer
│   ├── interfaces/        # Repository interfaces
│   ├── postgres/          # PostgreSQL implementation
│   ├── mysql/             # MySQL implementation
│   ├── mssql/             # MSSQL implementation
│   └── mongo/             # MongoDB implementation
├── security/               # Security utilities
│   ├── password/          # Password hashing
│   ├── jwt/               # JWT utilities
│   ├── totp/              # TOTP generation/validation
│   └── encryption/        # Encryption utilities
├── config/                 # Configuration
│   ├── loader/            # Config loading
│   └── validator/         # Config validation
└── internal/               # Internal utilities
    ├── cache/             # Caching layer
    ├── logger/            # Logging
    └── metrics/           # Metrics collection
```

## 🔧 Component Details

### 1. API Layer (`api/`)

**Purpose**: HTTP API interface, request/response handling

#### Structure

```
api/
├── handlers/
│   ├── auth_handler.go        # Authentication endpoints
│   ├── user_handler.go        # User management endpoints
│   ├── tenant_handler.go      # Tenant management endpoints
│   └── health_handler.go      # Health check endpoints
├── middleware/
│   ├── auth.go                # JWT authentication middleware
│   ├── rate_limit.go          # Rate limiting
│   ├── cors.go                # CORS handling
│   ├── logging.go             # Request logging
│   └── recovery.go            # Panic recovery
└── routes/
    └── routes.go              # Route definitions
```

#### Responsibilities

- HTTP request parsing and validation
- Response formatting
- Error handling
- Middleware orchestration
- Route registration

#### Key Interfaces

```go
type Handler interface {
    Handle(ctx *gin.Context)
}

type Middleware interface {
    Process(ctx *gin.Context) error
}
```

### 2. Auth Service (`auth/`)

**Purpose**: Authentication logic, MFA, token management

#### Structure

```
auth/
├── login/
│   ├── service.go            # Login service
│   └── validator.go          # Credential validator
├── mfa/
│   ├── totp.go               # TOTP implementation
│   ├── recovery.go           # Recovery codes
│   └── service.go            # MFA service
├── refresh/
│   └── service.go            # Token refresh service
├── hydra/
│   ├── client.go             # Hydra admin client
│   ├── login.go              # Login challenge handling
│   └── consent.go            # Consent handling
└── service.go                # Main auth service
```

#### Responsibilities

- Credential validation
- MFA verification
- Token refresh
- Hydra integration
- Session management (stateless)

#### Key Interfaces

```go
type AuthService interface {
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    VerifyMFA(ctx context.Context, req *MFARequest) (*MFAResponse, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
    Logout(ctx context.Context, token string) error
}

type HydraClient interface {
    AcceptLoginRequest(ctx context.Context, challenge string, subject string, claims map[string]interface{}) (*AcceptLoginResponse, error)
    GetLoginRequest(ctx context.Context, challenge string) (*LoginRequest, error)
    CreateOAuth2Client(ctx context.Context, client *OAuth2Client) error
}
```

### 3. Identity Service (`identity/`)

**Purpose**: User, tenant, group management

#### Structure

```
identity/
├── user/
│   ├── service.go            # User service
│   ├── repository.go         # User repository interface
│   └── model.go              # User model
├── tenant/
│   ├── service.go            # Tenant service
│   ├── repository.go         # Tenant repository interface
│   └── model.go              # Tenant model
├── group/
│   ├── service.go            # Group service
│   ├── repository.go         # Group repository interface
│   └── model.go              # Group model
└── credential/
    ├── service.go            # Credential service
    ├── repository.go         # Credential repository interface
    └── model.go              # Credential model
```

#### Responsibilities

- User CRUD operations
- Tenant management
- Group management
- Credential management
- User-tenant relationships

#### Key Interfaces

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    GetByUsername(ctx context.Context, username string, tenantID string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, tenantID string, filters *UserFilters) ([]*User, error)
}

type TenantRepository interface {
    Create(ctx context.Context, tenant *Tenant) error
    GetByID(ctx context.Context, id string) (*Tenant, error)
    GetByDomain(ctx context.Context, domain string) (*Tenant, error)
    Update(ctx context.Context, tenant *Tenant) error
    Delete(ctx context.Context, id string) error
}
```

### 4. Policy Service (`policy/`)

**Purpose**: Authorization, roles, permissions, claims

#### Structure

```
policy/
├── rbac/
│   ├── service.go            # RBAC service
│   ├── repository.go         # Role/permission repository
│   └── model.go              # Role/permission models
├── abac/
│   ├── service.go            # ABAC service
│   └── evaluator.go         # Attribute evaluator
├── claims/
│   ├── builder.go            # Claims builder
│   └── mapper.go             # Claims mapper
└── permissions/
    ├── service.go            # Permission service
    └── evaluator.go         # Permission evaluator
```

#### Responsibilities

- Role management
- Permission management
- Claims building
- Authorization decisions
- Policy evaluation

#### Key Interfaces

```go
type PolicyService interface {
    GetUserRoles(ctx context.Context, userID string, tenantID string) ([]*Role, error)
    GetUserPermissions(ctx context.Context, userID string, tenantID string) ([]string, error)
    BuildClaims(ctx context.Context, user *User, tenant *Tenant) (map[string]interface{}, error)
    Evaluate(ctx context.Context, userID string, resource string, action string) (bool, error)
}

type ClaimsBuilder interface {
    Build(ctx context.Context, user *User, tenant *Tenant, roles []*Role, permissions []string) (map[string]interface{}, error)
}
```

### 5. Storage Layer (`storage/`)

**Purpose**: Database abstraction, repository implementations

#### Structure

```
storage/
├── interfaces/
│   ├── user_repository.go    # User repository interface
│   ├── tenant_repository.go  # Tenant repository interface
│   ├── role_repository.go    # Role repository interface
│   └── credential_repository.go # Credential repository interface
├── postgres/
│   ├── user_repository.go    # PostgreSQL user implementation
│   ├── tenant_repository.go  # PostgreSQL tenant implementation
│   ├── connection.go         # Database connection
│   └── migrations/           # Database migrations
├── mysql/
│   └── ...                   # MySQL implementations
├── mssql/
│   └── ...                   # MSSQL implementations
└── mongo/
    └── ...                   # MongoDB implementations
```

#### Responsibilities

- Database connection management
- Repository implementations
- Query optimization
- Transaction management
- Migration handling

#### Key Interfaces

```go
type Repository interface {
    BeginTx(ctx context.Context) (Transaction, error)
}

type Transaction interface {
    Commit() error
    Rollback() error
}
```

### 6. Security Module (`security/`)

**Purpose**: Security utilities, password hashing, encryption

#### Structure

```
security/
├── password/
│   └── hasher.go             # Argon2id hasher
├── jwt/
│   ├── generator.go          # JWT generator
│   ├── validator.go          # JWT validator
│   └── jwks.go               # JWKS endpoint
├── totp/
│   ├── generator.go          # TOTP generator
│   └── validator.go          # TOTP validator
└── encryption/
    └── encryptor.go          # Encryption utilities
```

#### Responsibilities

- Password hashing (Argon2id)
- JWT generation and validation
- TOTP generation and validation
- Encryption/decryption
- Key management

#### Key Interfaces

```go
type PasswordHasher interface {
    Hash(password string) (string, error)
    Verify(password string, hash string) (bool, error)
}

type JWTGenerator interface {
    Generate(claims map[string]interface{}) (string, error)
    Validate(token string) (*Claims, error)
}

type TOTPGenerator interface {
    GenerateSecret() (string, error)
    GenerateQRCode(secret string, user string) ([]byte, error)
    Validate(secret string, code string) (bool, error)
}
```

### 7. Config Module (`config/`)

**Purpose**: Configuration loading and validation

#### Structure

```
config/
├── loader/
│   └── loader.go            # Config loader
├── validator/
│   └── validator.go         # Config validator
└── config.go                # Config struct
```

#### Responsibilities

- Environment variable loading
- Configuration file parsing
- Configuration validation
- Default value management

#### Configuration Structure

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Hydra    HydraConfig
    Security SecurityConfig
    Logging  LoggingConfig
}
```

### 8. Internal Utilities (`internal/`)

**Purpose**: Shared utilities, caching, logging, metrics

#### Structure

```
internal/
├── cache/
│   └── cache.go             # Redis cache wrapper
├── logger/
│   └── logger.go            # Structured logger
└── metrics/
    └── metrics.go           # Metrics collection
```

#### Responsibilities

- Caching abstraction
- Structured logging
- Metrics collection
- Common utilities

## 🔄 Component Interactions

### Login Flow Component Interaction

```
API Handler (auth_handler.go)
    ↓
Auth Service (auth/service.go)
    ↓
    ├──→ Identity Service (identity/user/service.go)
    │       └──→ User Repository (storage/postgres/user_repository.go)
    │
    ├──→ Credential Validator (auth/login/validator.go)
    │       └──→ Security (security/password/hasher.go)
    │
    ├──→ MFA Service (auth/mfa/service.go)
    │       └──→ Security (security/totp/validator.go)
    │
    ├──→ Policy Service (policy/service.go)
    │       └──→ Claims Builder (policy/claims/builder.go)
    │
    └──→ Hydra Client (auth/hydra/client.go)
            └──→ ORY Hydra Admin API
```

## 📊 Component Dependencies

```
api/ → auth/, identity/, policy/
auth/ → identity/, policy/, security/, storage/
identity/ → storage/
policy/ → storage/
storage/ → database driver
security/ → (standalone)
config/ → (standalone)
internal/ → (standalone utilities)
```

## 🧪 Testing Strategy per Component

### Unit Tests

- Each component has isolated unit tests
- Mock interfaces for dependencies
- Test coverage > 80%

### Integration Tests

- Component integration tests
- Database integration tests
- Hydra integration tests

### Contract Tests

- Repository interface contracts
- Service interface contracts
- API contract tests

## 📚 Related Documentation

- [Architecture Overview](./overview.md)
- [Data Flow](./data-flow.md)
- [Integration Patterns](./integration-patterns.md)

