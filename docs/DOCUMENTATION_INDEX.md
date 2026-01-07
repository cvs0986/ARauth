# 📚 Documentation Index

Quick reference guide to all documentation in Nuage Identity.

---

## 🎯 By Purpose

### 📊 Status & Progress
- **Current Status**: [`status/IMPLEMENTATION_STATUS.md`](status/IMPLEMENTATION_STATUS.md) - Overall implementation status (100% complete)
- **Completion Summary**: [`status/COMPLETION_STATUS.md`](status/COMPLETION_STATUS.md) - Hybrid auth completion summary
- **Secure Auth Status**: [`status/SECURE_AUTH_IMPLEMENTATION_STATUS.md`](status/SECURE_AUTH_IMPLEMENTATION_STATUS.md) - Secure auth recommendation implementation
- **Hybrid Auth Progress**: [`status/HYBRID_AUTH_IMPLEMENTATION.md`](status/HYBRID_AUTH_IMPLEMENTATION.md) - Hybrid auth implementation progress
- **Progress History**: [`status/progress/`](status/progress/) - Historical progress tracking
- **Fixes**: [`status/fixes/`](status/fixes/) - Fix documentation

### 🔐 Authentication
- **Flows Guide**: [`guides/authentication/AUTHENTICATION_FLOWS_GUIDE.md`](guides/authentication/AUTHENTICATION_FLOWS_GUIDE.md) - Complete guide for Direct JWT and OAuth2/OIDC
- **Recommendation**: [`architecture/authentication/SECURE_AUTH_RECOMMENDATION.md`](architecture/authentication/SECURE_AUTH_RECOMMENDATION.md) - Secure authentication recommendation
- **Flow Details**: [`security/authentication-flow-recommendation.md`](security/authentication-flow-recommendation.md) - Detailed authentication flow
- **Implementation Plan**: [`security/implementation-plan.md`](security/implementation-plan.md) - Security implementation plan
- **Token Configuration**: [`security/token-lifetime-configuration.md`](security/token-lifetime-configuration.md) - Token lifetime configuration

### 🧪 Testing
- **Testing Guide**: [`guides/testing/TESTING_GUIDE.md`](guides/testing/TESTING_GUIDE.md) - Comprehensive testing guide
- **Quick Start**: [`guides/testing/QUICK_START_TESTING.md`](guides/testing/QUICK_START_TESTING.md) - Quick start for testing
- **Code Coverage**: [`guides/testing/CODE_COVERAGE_GUIDE.md`](guides/testing/CODE_COVERAGE_GUIDE.md) - Code coverage guide
- **E2E Strategy**: [`testing/e2e-testing-strategy.md`](testing/e2e-testing-strategy.md) - E2E testing strategy
- **Integration Tests**: [`testing/integration-tests.md`](testing/integration-tests.md) - Integration tests
- **Performance**: [`testing/performance.md`](testing/performance.md) - Performance testing

### 🚀 Deployment
- **Using as IAM Service**: [`guides/deployment/using-as-iam-service.md`](guides/deployment/using-as-iam-service.md) - How to use Nuage Identity like Keycloak
- **Docker Compose**: [`deployment/docker-compose.md`](deployment/docker-compose.md) - Docker Compose setup
- **Kubernetes**: [`deployment/kubernetes.md`](deployment/kubernetes.md) - Kubernetes deployment
- **Production Guide**: [`deployment/production-guide.md`](deployment/production-guide.md) - Production deployment
- **Configuration**: [`deployment/configuration.md`](deployment/configuration.md) - Configuration guide
- **Monitoring**: [`deployment/monitoring.md`](deployment/monitoring.md) - Monitoring setup

### 🔗 Integration
- **Integration Guide**: [`guides/integration/integration-guide.md`](guides/integration/integration-guide.md) - Integration examples and patterns
- **Integration Patterns**: [`architecture/integration-patterns.md`](architecture/integration-patterns.md) - Integration patterns

### 🏗️ Architecture
- **Overview**: [`architecture/overview.md`](architecture/overview.md) - System architecture overview
- **Components**: [`architecture/components.md`](architecture/components.md) - Component architecture
- **Data Flow**: [`architecture/data-flow.md`](architecture/data-flow.md) - Data flow diagrams
- **Scalability**: [`architecture/scalability.md`](architecture/scalability.md) - Scalability architecture
- **Frontend-Backend**: [`architecture/frontend-backend-integration.md`](architecture/frontend-backend-integration.md) - Frontend-backend integration

### 🔒 Security
- **Security Technical**: [`technical/security.md`](technical/security.md) - Security technical details
- **Token Configuration**: [`security/token-lifetime-configuration.md`](security/token-lifetime-configuration.md) - Token lifetime configuration
- **Auth Flow**: [`security/authentication-flow-recommendation.md`](security/authentication-flow-recommendation.md) - Authentication flow recommendations

### 📋 Planning
- **Repository Structure**: [`planning/repository-structure-decision.md`](planning/repository-structure-decision.md) - Monorepo vs polyrepo decision
- **Branching Strategy**: [`planning/BRANCHES.md`](planning/BRANCHES.md) - Branching strategy
- **Roadmap**: [`planning/roadmap.md`](planning/roadmap.md) - Project roadmap
- **Timeline**: [`planning/timeline.md`](planning/timeline.md) - Project timeline
- **Strategy**: [`planning/strategy.md`](planning/strategy.md) - Project strategy
- **Risk Analysis**: [`planning/risk-analysis.md`](planning/risk-analysis.md) - Risk analysis

### 🔧 Technical
- **API Design**: [`technical/api-design.md`](technical/api-design.md) - API design specifications
- **Database Schema**: [`technical/database-design.md`](technical/database-design.md) - Database schema documentation
- **Tech Stack**: [`technical/tech-stack.md`](technical/tech-stack.md) - Technology stack
- **Testing Strategy**: [`technical/testing-strategy.md`](technical/testing-strategy.md) - Testing strategy

### ⚙️ Setup
- **GitHub Setup**: [`guides/setup/GITHUB_SETUP.md`](guides/setup/GITHUB_SETUP.md) - GitHub setup guide
- **Repository Setup**: [`guides/setup/SETUP_REPOSITORY.md`](guides/setup/SETUP_REPOSITORY.md) - Repository setup
- **Cursor Integration**: [`guides/setup/CURSOR_GITHUB_INTEGRATION.md`](guides/setup/CURSOR_GITHUB_INTEGRATION.md) - Cursor-GitHub integration

### 📡 API
- **API Reference**: [`api/README.md`](api/README.md) - API endpoints reference
- **OpenAPI Spec**: [`api/openapi.yaml`](api/openapi.yaml) - OpenAPI specification

---

## 🗂️ By Directory

### `status/` - Implementation Status
All status and progress tracking documents.
- Main status documents
- `progress/` - Historical progress
- `fixes/` - Fix documentation

### `guides/` - How-To Guides
Step-by-step guides organized by topic:
- `authentication/` - Authentication flow guides
- `testing/` - Testing guides
- `setup/` - Setup and configuration guides
- `deployment/` - Deployment guides
- `integration/` - Integration guides

### `architecture/` - Architecture Documents
High-level design and architectural decisions:
- `authentication/` - Authentication architecture
- System architecture
- Integration patterns

### `security/` - Security Documentation
Security-related documentation and recommendations.

### `planning/` - Planning Documents
Project planning and architectural decisions.

### `technical/` - Technical Documentation
Technical specifications and detailed documentation.

### `testing/` - Testing Documentation
Testing strategies, methodologies, and best practices.

### `deployment/` - Deployment Documentation
Deployment guides and production setup instructions.

### `api/` - API Documentation
API endpoint documentation and reference.

### `archive/` - Archived Documentation
Archived documentation for historical reference.

---

## 🔍 Quick Search

### "How do I test the system?"
→ [`guides/testing/TESTING_GUIDE.md`](guides/testing/TESTING_GUIDE.md)

### "What's the current implementation status?"
→ [`status/IMPLEMENTATION_STATUS.md`](status/IMPLEMENTATION_STATUS.md)

### "How does authentication work?"
→ [`guides/authentication/AUTHENTICATION_FLOWS_GUIDE.md`](guides/authentication/AUTHENTICATION_FLOWS_GUIDE.md)

### "How do I deploy this?"
→ [`guides/deployment/using-as-iam-service.md`](guides/deployment/using-as-iam-service.md)

### "What API endpoints are available?"
→ [`api/README.md`](api/README.md)

### "How do I configure token lifetimes?"
→ [`security/token-lifetime-configuration.md`](security/token-lifetime-configuration.md)

### "How do I set up GitHub integration?"
→ [`guides/setup/GITHUB_SETUP.md`](guides/setup/GITHUB_SETUP.md)

### "What's the architecture?"
→ [`architecture/overview.md`](architecture/overview.md)

---

## 📝 Document Maintenance

- **Status docs** are updated as implementation progresses
- **Guides** are updated when features change
- **Architecture docs** are updated when design decisions change
- **Technical docs** are updated when specifications change

---

**Last Updated**: 2026-01-08
