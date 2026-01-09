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

Short answer first, then the **clean, final model**:

> **Yes — some features and capabilities *must* exist at the user level, but only as enforcement and state, never as independent power.**

User-level controls **must not break** system or tenant guarantees.

---

# ✅ Final Principle (Very Important)

ARauth follows **strict downward inheritance**:

```
SYSTEM (maximum possible)
  → TENANT (what is enabled)
      → USER (what is enforced / enrolled)
```

👉 **Users never enable capabilities**
👉 **Users only comply with or consume capabilities**

---

# 🧠 Capability Levels — What Each Level Means

| Level               | Meaning                                       |
| ------------------- | --------------------------------------------- |
| **System**          | What exists at all (platform guardrails)      |
| **System → Tenant** | What this tenant is allowed to use            |
| **Tenant**          | What the tenant enables and enforces          |
| **User**            | What the user has enrolled in / is subject to |

Users are **never decision-makers** for security posture.

---

# 1️⃣ User-Level Capabilities: What SHOULD Exist

These are **stateful, per-user attributes**, not feature toggles.

## ✅ Authentication State (User Level)

| Capability     | User Level Role  | Why                  |
| -------------- | ---------------- | -------------------- |
| Password       | Has / resets     | Identity ownership   |
| MFA status     | Enrolled / not   | Security enforcement |
| TOTP           | Enrolled secrets | Per-user secret      |
| Recovery codes | Generated / used | Account recovery     |
| MFA bypass     | ❌ Not allowed    | Security             |

Example user state:

```json
{
  "user_id": "u-123",
  "mfa_enrolled": true,
  "totp_enrolled": true,
  "recovery_codes_remaining": 5
}
```

---

## ✅ Identity Constraints (User Level)

| Constraint       | Controlled By | Applied To User |
| ---------------- | ------------- | --------------- |
| Account disabled | Tenant        | Yes             |
| Account locked   | System        | Yes             |
| Password expired | Tenant        | Yes             |
| MFA required     | Tenant/System | Yes             |

User **cannot opt out**.

---

## ✅ Federation Identity (User Level)

| Feature       | User Level Meaning        |
| ------------- | ------------------------- |
| SAML login    | Linked external identity  |
| OIDC IdP      | Linked provider           |
| Multiple IdPs | Allowed if tenant enables |

Users don’t enable federation — they **link identities**.

---

# 2️⃣ What Users MUST NOT Control (Critical)

Users must **never** control:

| Feature             | Why                     |
| ------------------- | ----------------------- |
| Enable/disable MFA  | Weakens security        |
| Choose auth methods | Policy violation        |
| Token TTL           | Security risk           |
| OAuth grant types   | Protocol abuse          |
| Scopes              | Privilege escalation    |
| Roles               | Authorization violation |

❌ Even tenant admins cannot bypass this for themselves.

---

# 3️⃣ Feature-by-Feature: System vs Tenant vs User

## 🔐 MFA / TOTP

| Level         | Responsibility       |
| ------------- | -------------------- |
| System        | Allow MFA/TOTP       |
| System→Tenant | Permit tenant to use |
| Tenant        | Enable + enforce     |
| User          | Enroll & use         |

If tenant enforces MFA:

* User **must enroll**
* Login blocked until complete

---

## 🔑 OAuth2 / OIDC

| Level         | Responsibility             |
| ------------- | -------------------------- |
| System        | Allowed grants, PKCE rules |
| System→Tenant | Permit OIDC                |
| Tenant        | Create clients, scopes     |
| User          | Authenticate only          |

Users **never manage OAuth clients or scopes**.

---

## 🧬 Scopes & Permissions

| Level  | Responsibility            |
| ------ | ------------------------- |
| System | Namespace guardrails      |
| Tenant | Define scopes & roles     |
| User   | Receives scopes via roles |

Scopes are **derived**, never chosen.

---

## 🔁 Sessions & Tokens

| Level  | Responsibility    |
| ------ | ----------------- |
| System | Max TTL           |
| Tenant | TTL within limits |
| User   | Token usage       |

Users cannot extend sessions.

---

## 🏢 Tenant Membership

| Level  | Responsibility |
| ------ | -------------- |
| Tenant | Invite users   |
| User   | Accept invite  |
| System | No involvement |

---

# 4️⃣ Special Case: Tenant Admin Users

Even tenant admins are still **users**, not system actors.

They:

* Are subject to MFA
* Cannot disable their own MFA
* Cannot escalate to system roles
* Cannot change system limits

This avoids **self-privilege loopholes**.

---

# 5️⃣ User-Level Capability Evaluation (How IAM Enforces)

At login:

1. Resolve **system guardrails**
2. Resolve **tenant allowed capabilities**
3. Resolve **tenant enforcement rules**
4. Evaluate **user enrollment/state**
5. Enforce missing requirements (e.g., MFA enrollment)

Pseudo-logic:

```text
if tenant.mfa_required and not user.mfa_enrolled:
  block_login_and_force_enrollment()
```

---

# 6️⃣ Why This Model Is Correct (Industry Alignment)

| Platform | User-Level Control         |
| -------- | -------------------------- |
| AWS IAM  | Users comply, don’t decide |
| Azure AD | Users enroll MFA           |
| Okta     | Users enroll factors       |
| Auth0    | Users enroll MFA           |

No serious IAM lets users decide security posture.

---

# 🏁 Final Rules (Put This in Docs)

> **In ARauth, users do not enable security capabilities.
> They enroll in, comply with, and operate under capabilities enabled and enforced by the tenant and system.**

> **All security power flows downward; no capability flows upward.**

---

## ✅ One-line summary

* ✔ System defines limits
* ✔ Tenant enables & enforces
* ✔ User enrolls & complies
* ❌ User never weakens security

---

Below is a **single, clean inheritance diagram** that shows **how features & capabilities flow in ARauth** — **top-down only**, with **no upward overrides**.

This is the **canonical mental model** you should use everywhere (docs, code, reviews).

---

## 🔽 ARauth Capability Inheritance Diagram (Authoritative)

![Image](https://www.kuppingercole.com/pics/IAM_Reference_Architecture.jpg)

![Image](https://docs.aws.amazon.com/images/prescriptive-guidance/latest/patterns/images/pattern-img/4306bc76-22a7-45ca-a107-43df6c6f7ac8/images/700faf4d-c28f-4814-96aa-2d895cdcb518.png)

![Image](https://images.ctfassets.net/00voh0j35590/1Qe7iag3FfvvdyXWI4VzLU/32bc3a49e706b970ba351772102af9b4/IAM_diagram_2.png)

```
┌──────────────────────────────────────────────┐
│ SYSTEM (Platform / Control Plane)             │
│                                              │
│ • Defines WHAT EXISTS                         │
│ • Hard security guardrails                    │
│                                              │
│ Examples:                                    │
│ - MFA supported? (yes/no)                    │
│ - TOTP supported?                            │
│ - SAML supported?                            │
│ - OAuth2/OIDC supported?                     │
│ - Max token TTL                              │
│ - Allowed grant types                        │
│                                              │
│ ❌ Cannot act as tenant                      │
│ ❌ Cannot bypass tenant RBAC                 │
└───────────────────────┬──────────────────────┘
                        │
                        │ Allowed Capabilities
                        ▼
┌──────────────────────────────────────────────┐
│ SYSTEM → TENANT CAPABILITY ASSIGNMENT         │
│                                              │
│ • What THIS tenant is allowed to use          │
│ • Per-tenant feature flags                    │
│                                              │
│ Examples:                                    │
│ - Tenant A: MFA + OIDC + SAML                 │
│ - Tenant B: MFA + OIDC (no SAML)              │
│ - Tenant C: OIDC only                         │
│                                              │
│ Controlled ONLY by system admins              │
└───────────────────────┬──────────────────────┘
                        │
                        │ Enabled Features
                        ▼
┌──────────────────────────────────────────────┐
│ TENANT (Organization Plane)                   │
│                                              │
│ • Chooses what to ENABLE                      │
│ • Enforces security policies                  │
│                                              │
│ Examples:                                    │
│ - Enable MFA                                 │
│ - Require MFA for admins                     │
│ - Enable TOTP                                │
│ - Enable OAuth2/OIDC                         │
│ - Configure SAML IdP                         │
│ - Create roles, permissions, scopes           │
│                                              │
│ ❌ Cannot exceed system limits                │
│ ❌ Cannot weaken platform security            │
└───────────────────────┬──────────────────────┘
                        │
                        │ Enforcement & State
                        ▼
┌──────────────────────────────────────────────┐
│ USER (Identity Plane)                         │
│                                              │
│ • Enrolls & COMPLIES                          │
│ • Has state, not power                        │
│                                              │
│ Examples:                                    │
│ - Password set                               │
│ - TOTP enrolled                              │
│ - MFA completed                              │
│ - External IdP linked                        │
│                                              │
│ ❌ Cannot enable/disable features             │
│ ❌ Cannot choose scopes or roles              │
└──────────────────────────────────────────────┘
```

---

## 🔐 Key Rules (Non-Negotiable)

### 1️⃣ Direction is ONE-WAY

```
SYSTEM → TENANT → USER
```

There is **no reverse inheritance**.

---

### 2️⃣ Meaning of Each Level

| Level             | Meaning                     |
| ----------------- | --------------------------- |
| **System**        | What is even possible       |
| **System→Tenant** | What this tenant may use    |
| **Tenant**        | What is enabled & enforced  |
| **User**          | What is enrolled & required |

---

### 3️⃣ Capability vs State (Critical Distinction)

| Concept                              | Where it lives |
| ------------------------------------ | -------------- |
| Capability (MFA allowed?)            | System         |
| Permission (MFA allowed for tenant?) | System→Tenant  |
| Enforcement (MFA required?)          | Tenant         |
| Enrollment (TOTP secret)             | User           |

---

## 🔑 Example Walkthrough (Concrete)

### MFA Example

1. **System**

   * MFA supported = ✅

2. **System → Tenant**

   * Tenant allowed MFA = ✅

3. **Tenant**

   * MFA enabled
   * MFA required for admins

4. **User**

   * Must enroll TOTP
   * Cannot skip MFA
   * Cannot disable MFA

---

## 🪪 Token Reflection (Result of Inheritance)

Tokens only reflect **what actually happened**, not what is possible.

```json
{
  "tenant_id": "tenant-abc",
  "acr": "mfa",
  "amr": ["totp"]
}
```

---

## 🏁 One-Line Rule (Put This Everywhere)

> **In ARauth, capabilities flow strictly downward: the system defines limits, tenants enforce policies, and users comply through enrollment — with no upward overrides.**

---

