# Architecture Overview

## 🎯 System Purpose

ARauth Identity is a **headless Identity & Access Management (IAM) platform** that provides OAuth2/OIDC capabilities without a hosted login UI. Applications bring their own authentication UI and integrate with the IAM API to obtain tokens.

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Applications                        │
│  (Web, Mobile, SPA, Native Apps with Custom Login UI)           │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTPS / REST API
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      IAM API Service (Go)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Auth Service│  │ Identity Svc │  │ Policy Svc   │          │
│  │  - Login     │  │  - Users     │  │  - RBAC      │          │
│  │  - MFA       │  │  - Tenants   │  │  - ABAC      │          │
│  │  - Refresh   │  │  - Groups    │  │  - Policies  │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                 │                  │                  │
│         └─────────────────┴──────────────────┘                  │
│                           │                                      │
│                           │ Admin API                            │
└───────────────────────────┼──────────────────────────────────────┘
                            │
┌───────────────────────────▼──────────────────────────────────────┐
│                    ORY Hydra (OAuth2/OIDC)                        │
│  - Authorization Code + PKCE                                     │
│  - Client Credentials                                            │
│  - Token Issuance                                                │
│  - JWT Signing                                                   │
└───────────────────────────┬──────────────────────────────────────┘
                            │
┌───────────────────────────┴──────────────────────────────────────┐
│                         Data Layer                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  IAM DB      │  │  Hydra DB    │  │   Redis      │          │
│  │  (PostgreSQL)│  │  (PostgreSQL)│  │  - Sessions  │          │
│  │  - Users     │  │  - OAuth2    │  │  - OTP       │          │
│  │  - Tenants   │  │  - Clients   │  │  - Rate Limit│          │
│  │  - Roles     │  │  - Tokens    │  │              │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└──────────────────────────────────────────────────────────────────┘
```

## 🔑 Key Architectural Principles

### 1. Separation of Concerns

- **Hydra**: Pure OAuth2/OIDC provider, no business logic
- **IAM API**: User management, authentication, authorization logic
- **Client Apps**: Own their UI/UX, call IAM API for tokens

### 2. Stateless Design

- No server-side sessions
- All state in JWTs or external storage (Redis)
- Horizontally scalable

### 3. API-First

- RESTful API design
- OpenAPI specification
- No HTML/UI endpoints

### 4. Database Abstraction

- Repository pattern
- Interface-based design
- Support multiple databases (PostgreSQL, MySQL, MSSQL, MongoDB)

### 5. Security by Design

- Argon2id password hashing
- MFA support (TOTP)
- Rate limiting
- JWT with short expiration
- Refresh token rotation

## 🔄 Authentication Flow

### Standard Login Flow

```
1. Client App → IAM API: POST /auth/login
   { username, password, tenant_id }

2. IAM API:
   - Validates credentials
   - Checks tenant status
   - Enforces MFA (if enabled)
   - Builds claims

3. IAM API → Hydra Admin API: Accept Login Request
   - Injects custom claims
   - Gets authorization code

4. IAM API → Client App: Returns tokens
   { access_token, refresh_token, id_token }
```

### OAuth2 Authorization Code Flow (with PKCE)

```
1. Client App → Hydra: GET /oauth2/auth
   { client_id, redirect_uri, code_challenge, ... }

2. Hydra → IAM API: Login Challenge Callback
   { login_challenge }

3. IAM API → Client App: Returns login_challenge
   (Client app shows login UI)

4. Client App → IAM API: POST /auth/login
   { login_challenge, username, password }

5. IAM API → Hydra: Accept Login
   - Injects claims
   - Returns authorization code

6. Client App → Hydra: Exchange code for tokens
   { code, code_verifier, ... }

7. Hydra → Client App: Returns tokens
```

## 🧩 Core Components

### 1. IAM API Service
- **Framework**: Gin or Fiber (Go)
- **Responsibilities**:
  - HTTP API endpoints
  - Request validation
  - Middleware (rate limiting, CORS, logging)
  - Routing

### 2. Auth Service
- **Responsibilities**:
  - Credential validation
  - MFA verification
  - Token refresh
  - Logout
  - Hydra integration

### 3. Identity Service
- **Responsibilities**:
  - User management
  - Tenant management
  - Group management
  - Credential management

### 4. Policy Service
- **Responsibilities**:
  - Role management
  - Permission management
  - Claims building
  - Authorization decisions

### 5. Storage Layer
- **IAM Database**: Users, tenants, roles, permissions
- **Hydra Database**: OAuth2 clients, tokens, consent
- **Redis**: Sessions, OTP, rate limiting

## 🔐 Security Architecture

### Token Strategy

- **Access Tokens**: Short-lived (15 minutes), JWT format
- **Refresh Tokens**: Long-lived (7-30 days), opaque, rotated
- **ID Tokens**: OIDC standard, contains user info

### Claims Structure

```json
{
  "sub": "user-uuid",
  "tenant": "tenant-uuid",
  "roles": ["admin", "user"],
  "permissions": ["dc.read", "dc.write"],
  "acr": "mfa",
  "iss": "https://iam.example.com",
  "aud": "client-id",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### Password Security

- **Hashing**: Argon2id
- **Parameters**: Memory 64MB, Iterations 3, Parallelism 4
- **Salt**: Unique per password

### MFA

- **Method**: TOTP (Time-based One-Time Password)
- **Recovery Codes**: 10 single-use codes
- **Storage**: Encrypted in database

## 📊 Scalability Design

### Horizontal Scaling

- **Stateless API**: All instances identical
- **Load Balancing**: Round-robin or least connections
- **Database**: Connection pooling, read replicas
- **Redis**: Cluster mode for high availability

### Performance Targets

| Metric | Target |
|--------|--------|
| Startup Time | < 300ms |
| Login Latency | < 50ms |
| Token Issuance | < 10ms |
| Memory Usage | < 150MB per instance |
| Concurrent Logins | 10k+ |

### Caching Strategy

- **User Data**: Cache in Redis (TTL: 5 minutes)
- **Tenant Data**: Cache in memory (TTL: 10 minutes)
- **Role/Permissions**: Cache in memory (TTL: 15 minutes)
- **Rate Limits**: Redis with sliding window

## 🔌 Integration Points

### Client Applications

- **REST API**: Standard HTTP/HTTPS
- **OAuth2/OIDC**: Standard protocols
- **JWKS**: Public key endpoint for token validation

### Microservices

- **JWT Validation**: Self-contained, no IAM calls
- **Claims Extraction**: From JWT payload
- **Authorization**: Based on roles/permissions in token

### External Systems

- **LDAP/AD**: Future integration (optional)
- **SAML**: Future integration (optional)
- **Webhooks**: Event notifications (optional)

## 🚀 Deployment Architecture

### Kubernetes

```
┌─────────────────────────────────────────┐
│         Kubernetes Cluster              │
│                                         │
│  ┌──────────────┐  ┌──────────────┐   │
│  │  IAM API     │  │  IAM API     │   │
│  │  (Pod 1)     │  │  (Pod N)     │   │
│  └──────┬───────┘  └──────┬───────┘   │
│         │                 │            │
│         └────────┬────────┘            │
│                  │                      │
│         ┌────────▼────────┐            │
│         │  Service (LB)   │            │
│         └─────────────────┘            │
│                                         │
│  ┌──────────────┐  ┌──────────────┐   │
│  │  Hydra       │  │  Redis       │   │
│  │  (StatefulSet)│  │  (Cluster)   │   │
│  └──────────────┘  └──────────────┘   │
│                                         │
│  ┌──────────────┐  ┌──────────────┐   │
│  │  PostgreSQL  │  │  PostgreSQL  │   │
│  │  (IAM DB)    │  │  (Hydra DB)  │   │
│  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────┘
```

### On-Premise (Docker Compose)

- Single-node deployment
- All services in one compose file
- Suitable for small to medium deployments

## 📈 Monitoring & Observability

### Metrics

- Request latency (p50, p95, p99)
- Error rates
- Token issuance rate
- Active users
- MFA success/failure rates

### Logging

- Structured logging (JSON)
- Request/response logging
- Security event logging
- Error tracking

### Tracing

- Distributed tracing (OpenTelemetry)
- Request correlation IDs
- Performance profiling

## 🔄 Future Considerations

### Phase 2 Features

- LDAP/Active Directory integration
- SAML 2.0 support
- Social login (OAuth2 providers)
- Webhook events
- Audit logging

### Phase 3 Features

- Policy engine (OPA integration)
- Advanced ABAC
- Risk-based authentication
- Device fingerprinting

## 📚 Related Documentation

- [Components](./components.md) - Detailed component documentation
- [Data Flow](./data-flow.md) - Authentication flows
- [Integration Patterns](./integration-patterns.md) - Hydra integration
- [Scalability](./scalability.md) - Scalability design

