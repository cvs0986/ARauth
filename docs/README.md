# 📚 Nuage Identity Documentation

This directory contains all documentation for the Nuage Identity IAM system, organized by purpose and category.

---

## 📁 Directory Structure

```
docs/
├── README.md (this file)
│
├── status/                          # Implementation status and progress tracking
│   ├── IMPLEMENTATION_STATUS.md    # Current implementation status (100% complete)
│   ├── COMPLETION_STATUS.md        # Hybrid auth completion summary
│   ├── SECURE_AUTH_IMPLEMENTATION_STATUS.md  # Secure auth recommendation status
│   └── HYBRID_AUTH_IMPLEMENTATION.md         # Hybrid auth implementation progress
│
├── guides/                          # How-to guides and tutorials
│   ├── authentication/             # Authentication flow guides
│   │   └── AUTHENTICATION_FLOWS_GUIDE.md  # Complete guide for Direct JWT and OAuth2/OIDC flows
│   ├── testing/                    # Testing guides
│   │   ├── TESTING_GUIDE.md        # Comprehensive testing guide
│   │   ├── QUICK_START_TESTING.md  # Quick start for testing
│   │   └── CODE_COVERAGE_GUIDE.md  # Code coverage guide
│   ├── setup/                      # Setup and configuration guides
│   │   ├── GITHUB_SETUP.md
│   │   ├── SETUP_REPOSITORY.md
│   │   └── ... (other setup docs)
│   ├── deployment/                 # Deployment guides
│   │   └── using-as-iam-service.md # Using Nuage Identity as IAM service
│   └── integration/                # Integration guides
│       └── integration-guide.md    # Integration examples
│
├── architecture/                   # Architecture and design documents
│   └── authentication/             # Authentication architecture
│       └── SECURE_AUTH_RECOMMENDATION.md  # Secure authentication recommendation
│
├── security/                       # Security documentation
│   ├── authentication-flow-recommendation.md  # Detailed auth flow recommendations
│   ├── implementation-plan.md      # Security implementation plan
│   └── token-lifetime-configuration.md  # Token lifetime configuration guide
│
├── planning/                       # Planning and decisions
│   ├── repository-structure-decision.md  # Monorepo vs polyrepo decision
│   └── BRANCHES.md                # Branching strategy
│
├── technical/                      # Technical documentation
│   ├── api-design.md              # API design specifications
│   ├── security.md                # Security technical details
│   └── database-schema.md         # Database schema documentation
│
└── api/                           # API documentation
    └── README.md                  # API endpoints reference
```

---

## 🎯 Quick Navigation

### For Developers

- **Getting Started**: `guides/testing/QUICK_START_TESTING.md`
- **Authentication Flows**: `guides/authentication/AUTHENTICATION_FLOWS_GUIDE.md`
- **API Reference**: `api/README.md`
- **Testing**: `guides/testing/TESTING_GUIDE.md`

### For Architects

- **Architecture Overview**: `architecture/overview.md`
- **Security Architecture**: `security/authentication-flow-recommendation.md`
- **API Design**: `technical/api-design.md`

### For DevOps

- **Deployment Guide**: `guides/deployment/using-as-iam-service.md`
- **Configuration**: `security/token-lifetime-configuration.md`

### For Project Managers

- **Implementation Status**: `status/IMPLEMENTATION_STATUS.md`
- **Completion Status**: `status/COMPLETION_STATUS.md`

---

## 📖 Document Categories

### Status Documents (`status/`)
Track implementation progress, completion status, and what's been done.

### Guides (`guides/`)
Step-by-step instructions for common tasks:
- **Authentication**: How authentication flows work
- **Testing**: How to test the system
- **Deployment**: How to deploy and use the system
- **Integration**: How to integrate with other systems

### Architecture (`architecture/`)
High-level design decisions and architectural patterns.

### Security (`security/`)
Security-related documentation, recommendations, and implementation details.

### Planning (`planning/`)
Project planning documents and architectural decisions.

### Technical (`technical/`)
Technical specifications and detailed documentation.

### API (`api/`)
API endpoint documentation and reference.

---

## 🔍 Finding Documents

### By Purpose

- **"How do I...?"** → Check `guides/`
- **"What's the status?"** → Check `status/`
- **"How does X work?"** → Check `architecture/` or `technical/`
- **"How do I secure...?"** → Check `security/`
- **"What API endpoints exist?"** → Check `api/`

### By Topic

- **Authentication** → `guides/authentication/`, `architecture/authentication/`, `security/`
- **Testing** → `guides/testing/`
- **Deployment** → `guides/deployment/`
- **API** → `api/`, `technical/api-design.md`
- **Database** → `technical/database-schema.md`

---

## 📝 Document Naming Convention

- **Status docs**: `*_STATUS.md` or `*_IMPLEMENTATION.md`
- **Guides**: Descriptive names like `AUTHENTICATION_FLOWS_GUIDE.md`
- **Architecture**: `*_RECOMMENDATION.md` or `*_DESIGN.md`
- **Technical**: Topic-based names like `api-design.md`, `security.md`

---

## 🆕 Adding New Documentation

When adding new documentation:

1. **Status updates** → `docs/status/`
2. **How-to guides** → `docs/guides/<category>/`
3. **Architecture decisions** → `docs/architecture/`
4. **Security docs** → `docs/security/`
5. **Technical specs** → `docs/technical/`
6. **API docs** → `docs/api/`

Update this README when adding new major sections!

---

## 🔗 Related Documentation

- **Root README**: `../README.md` - Project overview
- **Frontend Status**: `../frontend/FRONTEND_STATUS.md` - Frontend implementation status

---

**Last Updated**: 2026-01-08
