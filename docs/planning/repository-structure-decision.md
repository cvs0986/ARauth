# Repository Structure Decision

## Question
Should frontend and IAM backend be in one repository (monorepo) or separate repositories?

## Analysis

### Option 1: Monorepo (Single Repository) ✅ **RECOMMENDED**

**Structure:**
```
nuage-identity/
├── cmd/server/          # IAM API backend
├── api/                 # API handlers
├── frontend/
│   ├── admin-dashboard/
│   ├── e2e-test-app/
│   └── shared/
├── docs/
└── ...
```

**Pros:**
- ✅ **Easier Development**: Single clone, single setup
- ✅ **Shared Code**: Types, constants, utilities can be shared
- ✅ **Atomic Commits**: Frontend and backend changes together
- ✅ **Simpler CI/CD**: One pipeline, coordinated deployments
- ✅ **Version Synchronization**: Same version for all components
- ✅ **Easier Testing**: Test full stack together
- ✅ **Better for Small Teams**: Less overhead
- ✅ **Unified Documentation**: All docs in one place

**Cons:**
- ⚠️ Larger repository size
- ⚠️ All developers need access to all code
- ⚠️ Can't deploy independently (but can be configured)

### Option 2: Separate Repositories

**Structure:**
```
nuage-identity-api/      # Backend repository
nuage-identity-admin/   # Admin dashboard repository
nuage-identity-e2e/      # E2E test app repository
```

**Pros:**
- ✅ Independent deployment
- ✅ Separate access control
- ✅ Smaller individual repos
- ✅ Different release cycles

**Cons:**
- ❌ More complex setup
- ❌ Harder to share code
- ❌ Version synchronization issues
- ❌ More CI/CD pipelines to maintain
- ❌ More overhead for small teams

## Recommendation: **Monorepo ✅**

### Reasoning

1. **Business Model**: 
   - IAM API is the core product
   - Frontend apps are admin/testing tools, not separate products
   - Users will primarily use the API (like Keycloak)

2. **Development Workflow**:
   - Frontend and backend are tightly coupled
   - API changes need frontend updates
   - Easier to develop and test together

3. **Distribution**:
   - Docker image contains only the API
   - Frontend can be optional/separate build
   - Users can use API without frontend

4. **Team Size**:
   - Small team = monorepo is better
   - Less overhead, easier coordination

5. **Shared Code**:
   - Types, constants, utilities benefit from sharing
   - Single source of truth

### Implementation

**Current Structure (Monorepo):**
```
nuage-identity/
├── cmd/server/          # IAM API
├── api/                 # API layer
├── frontend/            # Frontend apps
│   ├── admin-dashboard/
│   ├── e2e-test-app/
│   └── shared/
├── docs/                # Documentation
├── migrations/          # Database migrations
└── ...
```

**Benefits:**
- ✅ All code in one place
- ✅ Easy to navigate
- ✅ Shared types and constants
- ✅ Single CI/CD pipeline
- ✅ Unified versioning

### Deployment Strategy

Even with monorepo, we can deploy independently:

1. **Docker Images**:
   - `nuage-identity/iam-api:latest` - Backend only
   - `nuage-identity/admin-dashboard:latest` - Frontend only
   - Built from same repo, different Dockerfiles

2. **Kubernetes**:
   - Separate deployments
   - Independent scaling
   - Same repo, different manifests

3. **CI/CD**:
   - Build backend and frontend separately
   - Deploy independently
   - But from same repository

### When to Split?

Consider splitting if:
- Frontend becomes a separate product
- Different teams own frontend/backend
- Different release cycles needed
- Repository becomes too large (>1GB)
- Access control requirements differ significantly

## Decision: **Monorepo** ✅

**Action Items:**
- [x] Keep current monorepo structure
- [x] Frontend in `frontend/` directory
- [x] Shared code in `frontend/shared/`
- [ ] Update CI/CD for separate builds
- [ ] Document deployment strategy

---

**Decision Date**: 2024  
**Status**: Approved - Monorepo Structure

