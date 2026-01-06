
*(Headless IAM using ORY Hydra, No Login UI)*

---

## 🎯 Objective

Build a **lightweight, headless Identity & Access Management (IAM) platform** similar to Keycloak/Auth0, but:

* **NO hosted login UI**
* **Apps bring their own login UI**
* **OAuth2 / OIDC powered by ORY Hydra under the hood**
* **IAM is API-first**
* **Stateless & horizontally scalable**
* **DB-agnostic**
* **Deployable on Kubernetes, on-prem, or SaaS**

The system must be **production-grade**, **secure**, and **enterprise-ready**.

---

## 🧠 Architectural Principles

1. **Hydra is ONLY OAuth2/OIDC**

   * No users
   * No passwords
   * No UI
   * No business logic

2. **IAM owns**

   * Users
   * Credentials
   * MFA
   * Tenants
   * Roles / Permissions
   * Claims
   * Policies

3. **Apps own**

   * Login UI
   * Branding
   * UX

4. **Microservices NEVER call IAM**

   * They only validate JWTs

---

## 🏗️ High-Level Architecture

```
Client App (Web/Mobile)
 └── Custom Login UI
       └── IAM API (/auth/login)
             ├── Identity Service
             ├── Credential Validation
             ├── MFA (optional)
             ├── Claims Builder
             └── ORY Hydra Admin API
                    └── OAuth2 / OIDC Tokens
```

---

## 🧩 Components to Build

### 1️⃣ IAM API (Core)

**Language:** Go
**Framework:** Gin or Fiber
**Auth:** OAuth2 / JWT
**Stateless:** Yes

#### Responsibilities

* Authenticate users
* Perform MFA
* Call Hydra Admin APIs
* Inject JWT claims
* Manage tenants, roles, permissions

---

### 2️⃣ Identity Service

#### Entities

* Tenant
* User
* Group
* Role
* Permission
* Credential

#### Interfaces (DB Agnostic)

```go
type UserRepository interface {
  Create(user *User) error
  GetByUsername(username string) (*User, error)
}
```

👉 NO SQL in business logic

---

### 3️⃣ Auth Service (Headless)

#### APIs

```
POST /auth/login
POST /auth/mfa/verify
POST /auth/refresh
POST /auth/logout
```

#### Login Flow

1. Validate username/password
2. Check tenant status
3. Enforce MFA (if enabled)
4. Build claims
5. Call Hydra to issue tokens
6. Return tokens to app

---

### 4️⃣ OAuth2 / OIDC (ORY Hydra)

#### Requirements

* Authorization Code + PKCE
* Client Credentials
* Refresh Token rotation
* JWT access tokens
* No UI
* No users

#### Hydra Integration

* Use **login_challenge**
* Use **accept login API**
* Inject custom claims

Hydra is **never exposed directly** to clients.

---

### 5️⃣ Claims Strategy

JWT **must include**:

```json
{
  "sub": "user-id",
  "tenant": "tenant-id",
  "roles": ["admin"],
  "permissions": ["dc.read", "dc.write"],
  "acr": "mfa",
  "iss": "your-iam"
}
```

---

### 6️⃣ Authorization Model

* RBAC (roles → permissions)
* ABAC (attributes)
* Policy-ready (OPA compatible)

IAM decides **what goes into the token**
Services decide **what to allow**

---

### 7️⃣ Security Requirements

Mandatory:

* Argon2id password hashing
* Rate limiting
* MFA (TOTP + recovery codes)
* Refresh token rotation
* Short-lived access tokens
* Key rotation via JWKS

---

## 🗄️ Storage Requirements

### IAM DB

* PostgreSQL (default)
* Must support adapters for:

  * MySQL
  * MSSQL
  * MongoDB

### Redis

* OTP
* Login sessions
* Rate limits

### Hydra DB

* Separate DB
* OAuth2 only

---

## 🧪 Non-Functional Requirements

| Area              | Target  |
| ----------------- | ------- |
| Startup time      | < 300ms |
| Login latency     | < 50ms  |
| Token issuance    | < 10ms  |
| Memory            | < 150MB |
| Concurrent logins | 10k+    |

---

## 🐳 Deployment

### Kubernetes

* Helm charts
* HPA enabled
* Stateless pods
* Config via env

### On-Prem

* Docker Compose
* Single node support

---

## 📁 Project Structure (Expected)

```
iam/
 ├── cmd/
 ├── api/
 ├── auth/
 ├── identity/
 ├── policy/
 ├── hydra/
 ├── storage/
 │    ├── postgres/
 │    ├── mysql/
 │    └── mongo/
 ├── config/
 ├── security/
 └── main.go
```

---

## 🔍 What NOT to Build

❌ No login UI
❌ No HTML pages
❌ No sessions
❌ No server-side auth state
❌ No direct DB access from handlers

---

## 🧪 Tests Required

* Unit tests for:

  * Credential validation
  * Token issuance
  * Claims generation
* Integration tests with Hydra
* JWT validation tests

---

## 📦 Deliverables

1. IAM API service
2. Hydra integration
3. DB abstraction layer
4. Helm charts
5. OpenAPI spec
6. Sample app integration
7. Security documentation

---

## 🏁 Final Constraint

**The system must allow any enterprise or app to bring its own login UI without breaking OAuth2/OIDC compliance.**

---

### 🚀 End of Prompt
