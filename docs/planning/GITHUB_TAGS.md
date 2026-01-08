# GitHub Tags Structure for Capability Model Implementation

This document defines the tag structure to be used for all GitHub issues and pull requests related to the capability model implementation.

---

## Priority Tags

### `p0` - Critical
Must be completed first. Blocks other work or is required for core functionality.

**Usage**: All Phase 1-5 issues, critical security features, database migrations

### `p1` - Important
Should be completed soon but doesn't block other work.

**Usage**: Phase 6-7 issues, UI enhancements, documentation

### `p2` - Nice to Have
Can be deferred without impacting core functionality.

**Usage**: Future enhancements, optional features

---

## Component Tags

### `backend`
Backend/Go code changes (services, repositories, handlers, middleware)

**Examples**: #006, #007, #008, #009, #021, #022, #023

### `frontend`
Frontend/React code changes (components, pages, UI)

**Examples**: #014, #015, #016, #017, #018, #019, #020

### `database`
Database schema changes, migrations

**Examples**: #001, #002, #003, #004, #028

### `api`
API endpoint changes (new endpoints, modifications)

**Examples**: #010, #011, #012, #013

### `testing`
Test-related work (unit, integration, E2E tests)

**Examples**: #024, #025, #026

### `documentation`
Documentation updates and creation

**Examples**: #027

---

## Feature Tags

### `capability-model`
Core capability model implementation (all issues should have this)

**Usage**: All issues related to the capability model

### `system`
System-level features (system admin functionality)

**Examples**: #010, #011, #014, #015

### `tenant`
Tenant-level features (tenant admin functionality)

**Examples**: #012, #016

### `user`
User-level features (user enrollment, state)

**Examples**: #013, #017

### `mfa`
MFA/TOTP related features

**Usage**: Issues specifically about MFA/TOTP capabilities

### `oauth`
OAuth2/OIDC related features

**Usage**: Issues specifically about OAuth/OIDC capabilities

### `saml`
SAML related features

**Usage**: Issues specifically about SAML capabilities

### `security`
Security-related features and enforcement

**Examples**: #008, #009, #021, #022

---

## Type Tags

### `migration`
Database migration work

**Examples**: #001, #002, #003, #004, #028

### `service`
Service layer implementation

**Examples**: #006

### `repository`
Repository layer implementation

**Examples**: #007

### `middleware`
Middleware implementation

**Examples**: #021

### `ui`
UI component work

**Examples**: #014, #015, #016, #017, #018, #019, #020

### `integration`
Integration work (connecting components)

**Usage**: Issues that involve integrating multiple components

---

## Phase Tags

### `phase-1`
Phase 1: Database & Models

**Examples**: #001, #002, #003, #004, #005

### `phase-2`
Phase 2: Backend Core Logic

**Examples**: #006, #007, #008, #009

### `phase-3`
Phase 3: API Endpoints

**Examples**: #010, #011, #012, #013

### `phase-4`
Phase 4: Frontend Admin Dashboard

**Examples**: #014, #015, #016, #017, #018, #019, #020

### `phase-5`
Phase 5: Enforcement & Validation

**Examples**: #021, #022, #023

### `phase-6`
Phase 6: Testing & Documentation

**Examples**: #024, #025, #026, #027

### `phase-7`
Phase 7: Migration & Deployment

**Examples**: #028, #029, #030

---

## Tag Combinations

Issues typically have multiple tags. Common combinations:

### Database Migration Issue
- `database`, `migration`, `p0`, `capability-model`, `phase-1`

### Backend Service Issue
- `backend`, `service`, `p0`, `capability-model`, `phase-2`

### Frontend UI Issue
- `frontend`, `ui`, `p0`, `capability-model`, `phase-4`, `system` (or `tenant` or `user`)

### API Endpoint Issue
- `api`, `backend`, `p0`, `capability-model`, `phase-3`, `system` (or `tenant` or `user`)

### Testing Issue
- `testing`, `unit` (or `integration` or `e2e`), `p1`, `capability-model`, `phase-6`

---

## Creating Tags in GitHub

### Using GitHub CLI (gh)

```bash
# Create all tags
gh label create "p0" --description "Critical priority" --color "d73a4a"
gh label create "p1" --description "Important priority" --color "fbca04"
gh label create "p2" --description "Nice to have" --color "0e8a16"

gh label create "backend" --description "Backend code changes" --color "1d76db"
gh label create "frontend" --description "Frontend code changes" --color "bfd4f2"
gh label create "database" --description "Database changes" --color "5319e7"
gh label create "api" --description "API endpoint changes" --color "c2e0c6"
gh label create "testing" --description "Test-related work" --color "f9d0c4"
gh label create "documentation" --description "Documentation updates" --color "d4c5f9"

gh label create "capability-model" --description "Core capability model" --color "b60205"
gh label create "system" --description "System-level features" --color "0e8a16"
gh label create "tenant" --description "Tenant-level features" --color "1d76db"
gh label create "user" --description "User-level features" --color "fbca04"
gh label create "mfa" --description "MFA/TOTP features" --color "d73a4a"
gh label create "oauth" --description "OAuth2/OIDC features" --color "0052cc"
gh label create "saml" --description "SAML features" --color "5319e7"
gh label create "security" --description "Security-related" --color "b60205"

gh label create "migration" --description "Database migration" --color "5319e7"
gh label create "service" --description "Service layer" --color "1d76db"
gh label create "repository" --description "Repository layer" --color "0e8a16"
gh label create "middleware" --description "Middleware" --color "fbca04"
gh label create "ui" --description "UI component" --color "bfd4f2"
gh label create "integration" --description "Integration work" --color "c2e0c6"

gh label create "phase-1" --description "Phase 1: Database & Models" --color "d73a4a"
gh label create "phase-2" --description "Phase 2: Backend Core Logic" --color "fbca04"
gh label create "phase-3" --description "Phase 3: API Endpoints" --color "0e8a16"
gh label create "phase-4" --description "Phase 4: Frontend Admin Dashboard" --color "1d76db"
gh label create "phase-5" --description "Phase 5: Enforcement & Validation" --color "5319e7"
gh label create "phase-6" --description "Phase 6: Testing & Documentation" --color "bfd4f2"
gh label create "phase-7" --description "Phase 7: Migration & Deployment" --color "c2e0c6"
```

### Using GitHub Web UI

1. Go to repository → Labels
2. Click "New label"
3. Enter label name, description, and color
4. Click "Create label"
5. Repeat for all tags

---

## Tag Maintenance

- **Review quarterly**: Remove unused tags, add new ones as needed
- **Keep consistent**: Use existing tags before creating new ones
- **Document changes**: Update this document when tags are added/removed

---

## Tag Usage Guidelines

1. **Always include**: `capability-model` for all related issues
2. **Include phase tag**: One of `phase-1` through `phase-7`
3. **Include priority**: One of `p0`, `p1`, `p2`
4. **Include component**: One or more of `backend`, `frontend`, `database`, `api`, `testing`, `documentation`
5. **Include feature**: If applicable, one of `system`, `tenant`, `user`, `mfa`, `oauth`, `saml`, `security`
6. **Include type**: If applicable, one of `migration`, `service`, `repository`, `middleware`, `ui`, `integration`

**Example**: Issue #006 would have tags:
- `backend`, `service`, `p0`, `capability-model`, `phase-2`

