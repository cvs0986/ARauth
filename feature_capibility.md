Below is a **single, coherent, high-level design document** for **ARauth** that you can treat as a **source of truth**.
It clearly defines **boundaries**, **responsibilities**, **roles**, **scopes**, **bootstrap**, and **all major capabilities** (MFA, TOTP, OAuth2/OIDC, SAML, etc.).

This is written in a way that:

* You can put it in your **docs repo**
* You can give it to **Cursor / engineers**
* It will survive **security reviews, audits, and enterprise scrutiny**

---

# ARauth — System & Tenant Boundary Architecture

## 1. Core Philosophy

ARauth is built on **strict separation of responsibility** between:

1. **System (Platform / Control Plane)**
2. **Tenant (Organization / Customer Plane)**
3. **User (Identity Plane)**

> **System defines what is possible**
> **Tenant defines what is enabled**
> **User operates within enforced policies**

No shortcuts. No hidden privileges.

---

## 2. Planes of Control (Foundational Model)

```
┌──────────────────────────────────┐
│ SYSTEM CONTROL PLANE              │
│ (Platform / Master Admin)         │
│                                  │
│ - Tenant lifecycle                │
│ - Global security guardrails      │
│ - Platform roles & policies       │
└───────────────┬──────────────────┘
                │
┌───────────────▼──────────────────┐
│ TENANT PLANE                      │
│ (Organization / Customer)         │
│                                  │
│ - Users & groups                  │
│ - Tenant roles & permissions      │
│ - OAuth clients                   │
│ - MFA / SAML / OIDC config        │
└───────────────┬──────────────────┘
                │
┌───────────────▼──────────────────┐
│ USER PLANE                        │
│                                  │
│ - Login                           │
│ - MFA enrollment                  │
│ - Password / TOTP / SSO usage     │
└──────────────────────────────────┘
```

---

## 3. System (Platform) Responsibilities

### 3.1 What the System **CAN DO**

#### Tenant lifecycle

* Create tenant
* Suspend / resume tenant
* Delete tenant (with retention)
* Assign initial tenant owner
* View tenant configuration (read-only)

#### Platform security guardrails

* Define **minimum password policy**
* Define **maximum token lifetimes**
* Allow / disallow auth methods globally:

  * Password
  * MFA
  * TOTP
  * OAuth2/OIDC
  * SAML
* Enforce MFA for:

  * All tenants
  * Tenant admins
* Enforce PKCE, disallow unsafe grants

#### Governance

* View tenant audit logs
* Define audit retention
* View platform-wide metrics

---

### 3.2 What the System **CANNOT DO (by default)**

* ❌ Create normal users inside tenants
* ❌ Act as a tenant admin
* ❌ Assign tenant roles
* ❌ Access tenant application data
* ❌ Bypass tenant RBAC

**System manages tenants as objects, not as environments it “enters.”**

---

## 4. System Users, Roles & Permissions

### 4.1 System Users

System users:

* Do **not** belong to any tenant
* Authenticate against **system auth policy**
* Always require **strong MFA**

#### Principal type

```
principal_type = SYSTEM
tenant_id = null
```

---

### 4.2 System Roles (Predefined, Non-Editable)

System roles are **predefined and immutable**.

| Role           | Responsibility               |
| -------------- | ---------------------------- |
| system_owner   | Full platform control        |
| system_admin   | Tenant & platform management |
| system_auditor | Read-only governance         |

System roles **cannot be created or modified by tenants**.

---

### 4.3 Can a System Owner Create Other System Users?

✅ **YES — explicitly and intentionally**

A system owner can:

* Create other system users
* Assign them **specific system roles**
* Restrict responsibilities (least privilege)

Example:

* One system admin for billing
* One system auditor for compliance

All actions are:

* MFA protected
* Fully audited

---

## 5. Tenant Responsibilities

### 5.1 What Tenants **CAN DO**

#### User & access management

* Create / disable / delete users
* Assign tenant roles
* Manage groups
* Reset passwords
* Enforce MFA for users/admins

#### Tenant roles & permissions

* Create custom roles
* Create custom permissions
* Map roles → permissions
* Map permissions → scopes

#### OAuth2 / OIDC

* Enable OAuth2/OIDC (if allowed by system)
* Create OAuth clients
* Configure redirect URIs
* Choose allowed grant types (within system limits)
* Define scopes

#### Federation

* Configure OIDC federation
* Configure SAML IdPs
* Configure LDAP/AD (if enabled)

---

### 5.2 What Tenants **CANNOT DO**

* ❌ Create or manage other tenants
* ❌ Disable system-mandated security (e.g., MFA)
* ❌ Change signing algorithms
* ❌ Exceed system token limits
* ❌ Access other tenants

---

## 6. Tenant Roles, Permissions & Scopes

### 6.1 Role & Permission Model

**Tenant-defined, fully customizable**

```
Role
 └── Permissions
      └── Scopes
```

Example:

```
Role: tenant_admin
Permissions:
 - user.manage
 - client.manage
 - policy.manage
Scopes:
 - users:write
 - clients:write
```

---

### 6.2 Scope Model (Important)

#### System level

* Defines **allowed scope namespaces**
* Example:

  * `openid`
  * `profile`
  * `users.*`
  * `clients.*`

#### Tenant level

* Creates scopes within allowed namespaces
* Assigns scopes to roles

#### Token issuance

* IAM calculates **granted scopes**
* Hydra issues token
* Services enforce locally

---

## 7. Authentication & Security Features

### 7.1 Password Authentication

| Level  | Control         |
| ------ | --------------- |
| System | Minimum policy  |
| Tenant | Enable/disable  |
| User   | Change password |

Tenants cannot weaken system policy.

---

### 7.2 MFA & TOTP

#### System

* Allows / disallows MFA & TOTP
* Can enforce MFA globally or for admins

#### Tenant

* Enables MFA
* Chooses enforcement rules
* Chooses allowed methods (TOTP, later WebAuthn)

#### User

* Enrolls TOTP
* Manages recovery codes

MFA **never bypassable**.

---

### 7.3 OAuth2 / OIDC

* Protocol handled by **Hydra**
* Policy handled by **ARauth**

| Level  | Responsibility             |
| ------ | -------------------------- |
| System | Allowed grants, PKCE rules |
| Tenant | Clients, scopes, TTL       |
| User   | Login                      |

---

### 7.4 SAML

| Level  | Responsibility               |
| ------ | ---------------------------- |
| System | Enable/disable SAML globally |
| Tenant | Configure IdP, mappings      |
| User   | Login via SAML               |

SAML users still obey tenant MFA rules unless explicitly exempted.

---

## 8. Bootstrap & First Tenant User

### 8.1 System Bootstrap (Mandatory)

On first deployment:

* System is **uninitialized**
* No tenants
* No users

Bootstrap creates:

* **First system_owner**
* Marks system initialized

---

### 8.2 Tenant Creation Flow

Tenant can be created by:

* System admin (most common)
* Migration job

When tenant is created:

* **Exactly ONE tenant owner user is created**
* Assigned role: `tenant_owner` (predefined)
* Forced password setup
* Forced MFA enrollment

This is the **only time system creates a tenant user** by default.

---

### 8.3 Tenant Owner

Tenant owner:

* Full control of tenant
* Can create other tenant admins
* Cannot escalate to system roles

---

## 9. Enforcement Rules (Non-Negotiable)

* System tokens **cannot** call tenant APIs
* Tenant tokens **cannot** call system APIs
* Roles never cross planes
* Scopes are always reduced to least privilege
* All sensitive actions are audited

---

## 10. One-Page Summary (Put This in Docs)

```
SYSTEM
 - Creates tenants
 - Defines security boundaries
 - Manages platform users
 - Cannot manage tenant users (except bootstrap/migration)

TENANT
 - Manages users, roles, scopes
 - Configures MFA, OAuth2, SAML
 - Cannot bypass system rules

USER
 - Authenticates
 - Enrolls MFA
 - Uses granted access only

IAM ENFORCES EVERYTHING
HYDRA ISSUES TOKENS ONLY
SERVICES TRUST TOKENS
```

---

**Yes — and this is the *correct*, enterprise-grade design.**
In **ARauth**, the **system must be able to enable/disable features *per tenant***, *in addition to* global guardrails.

Below is a **clear, non-ambiguous model** that you can adopt as final.

---

# ✅ Final Decision

> **ARauth supports feature and capability control at THREE levels**
> **System (global) → System per-tenant → Tenant enablement**

This gives:

* Strong platform governance
* Tenant isolation
* Commercial flexibility
* Zero security ambiguity

---

# 🧠 The Three-Layer Capability Model

```
GLOBAL SYSTEM (What exists at all)
   ↓
SYSTEM → TENANT OVERRIDES (What this tenant may use)
   ↓
TENANT CONFIG (What the tenant actually enables)
```

Think: **Allowed → Permitted → Enabled**

---

# 1️⃣ Global System Level (Platform Guardrails)

**Purpose:** Define *what ARauth supports at all*.

### Examples

* MFA supported: ✅
* TOTP supported: ✅
* SAML supported: ❌ (disabled platform-wide)
* OAuth2/OIDC supported: ✅
* Max token TTL: 15 min
* PKCE mandatory: ✅

🔒 These are **hard limits**.
Tenants cannot exceed them.

---

# 2️⃣ System → Tenant Capability Assignment (Very Important)

**Purpose:** Decide *what a specific tenant is allowed to use*.

This enables:

* Tiered plans (Free / Pro / Enterprise)
* Risk-based restrictions
* Regulatory compliance

### Example per-tenant capability policy

```json
{
  "tenant_id": "tenant-abc",
  "allowed_capabilities": {
    "mfa": true,
    "totp": true,
    "saml": false,
    "oidc": true,
    "passwordless": false,
    "max_token_ttl": "10m"
  }
}
```

📌 **This is controlled ONLY by system admins.**

---

# 3️⃣ Tenant Enablement (Tenant Choice)

**Purpose:** Allow tenants to choose what to actually turn on.

Example:

```json
{
  "mfa": {
    "enabled": true,
    "required_for_admins": true
  },
  "oidc": {
    "enabled": true
  }
}
```

Rules:

* Tenant can enable **only what system allowed**
* Tenant can further restrict, never loosen

---

# 🧱 Capability Matrix (Clear & Explicit)

| Feature          | Global System  | System→Tenant    | Tenant           |
| ---------------- | -------------- | ---------------- | ---------------- |
| Password auth    | Enable/disable | Allow per tenant | Enable           |
| MFA              | Enable/disable | Allow per tenant | Enable + enforce |
| TOTP             | Enable/disable | Allow per tenant | Enable           |
| OAuth2/OIDC      | Enable/disable | Allow per tenant | Enable           |
| SAML             | Enable/disable | Allow per tenant | Enable           |
| Passwordless     | Enable/disable | Allow per tenant | Enable           |
| Max token TTL    | Set max        | Set per tenant   | Choose within    |
| Grant types      | Allowed list   | Allowed subset   | Enable subset    |
| Scope namespaces | Allowed list   | Allowed subset   | Create scopes    |

---

# 🔐 Enforcement Points (Where Logic Lives)

| Layer    | Responsibility                     |
| -------- | ---------------------------------- |
| System   | Validate requested configuration   |
| IAM Core | Enforce capability checks          |
| Hydra    | Issue tokens based on IAM decision |
| Services | Enforce scopes                     |

No logic duplication.

---

# 🧪 Real-World Alignment (Sanity Check)

| Platform | Per-Tenant Feature Control |
| -------- | -------------------------- |
| Auth0    | ✅                          |
| Okta     | ✅                          |
| Azure AD | ✅                          |
| AWS IAM  | ✅                          |

If ARauth didn’t support this, enterprises would reject it.

---

# 🪪 Token Reflection (Optional but Recommended)

Tokens may include **capability context** for auditing:

```json
{
  "tenant_id": "tenant-abc",
  "capabilities": {
    "mfa": true,
    "saml": false
  }
}
```

(Informational only, not authoritative.)

---

This gives you:

* Strong security guarantees
* Clean SaaS monetization
* Zero ambiguity
* Clean enforcement logic

---

