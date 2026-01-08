#!/bin/bash

# Script to manage GitHub issues, tags, and project board for Capability Model Implementation
# Requires GitHub CLI (gh) to be installed and authenticated

set -e

REPO="cvs0986/ARauth"  # GitHub repository
PROJECT_NUMBER=""  # Will be set after project is created/retrieved

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}GitHub Capability Model Issue Management${NC}"
echo "=========================================="

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed.${NC}"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub CLI.${NC}"
    echo "Run: gh auth login"
    exit 1
fi

# Function to create all tags
create_tags() {
    echo -e "${YELLOW}Creating GitHub tags...${NC}"
    
    # Priority tags
    gh label create "p0" --description "Critical priority" --color "d73a4a" --force 2>/dev/null || true
    gh label create "p1" --description "Important priority" --color "fbca04" --force 2>/dev/null || true
    gh label create "p2" --description "Nice to have" --color "0e8a16" --force 2>/dev/null || true
    
    # Component tags
    gh label create "backend" --description "Backend code changes" --color "1d76db" --force 2>/dev/null || true
    gh label create "frontend" --description "Frontend code changes" --color "bfd4f2" --force 2>/dev/null || true
    gh label create "database" --description "Database changes" --color "5319e7" --force 2>/dev/null || true
    gh label create "api" --description "API endpoint changes" --color "c2e0c6" --force 2>/dev/null || true
    gh label create "testing" --description "Test-related work" --color "f9d0c4" --force 2>/dev/null || true
    gh label create "documentation" --description "Documentation updates" --color "d4c5f9" --force 2>/dev/null || true
    
    # Feature tags
    gh label create "capability-model" --description "Core capability model" --color "b60205" --force 2>/dev/null || true
    gh label create "system" --description "System-level features" --color "0e8a16" --force 2>/dev/null || true
    gh label create "tenant" --description "Tenant-level features" --color "1d76db" --force 2>/dev/null || true
    gh label create "user" --description "User-level features" --color "fbca04" --force 2>/dev/null || true
    gh label create "mfa" --description "MFA/TOTP features" --color "d73a4a" --force 2>/dev/null || true
    gh label create "oauth" --description "OAuth2/OIDC features" --color "0052cc" --force 2>/dev/null || true
    gh label create "saml" --description "SAML features" --color "5319e7" --force 2>/dev/null || true
    gh label create "security" --description "Security-related" --color "b60205" --force 2>/dev/null || true
    
    # Type tags
    gh label create "migration" --description "Database migration" --color "5319e7" --force 2>/dev/null || true
    gh label create "service" --description "Service layer" --color "1d76db" --force 2>/dev/null || true
    gh label create "repository" --description "Repository layer" --color "0e8a16" --force 2>/dev/null || true
    gh label create "middleware" --description "Middleware" --color "fbca04" --force 2>/dev/null || true
    gh label create "ui" --description "UI component" --color "bfd4f2" --force 2>/dev/null || true
    gh label create "integration" --description "Integration work" --color "c2e0c6" --force 2>/dev/null || true
    
    # Phase tags
    gh label create "phase-1" --description "Phase 1: Database & Models" --color "d73a4a" --force 2>/dev/null || true
    gh label create "phase-2" --description "Phase 2: Backend Core Logic" --color "fbca04" --force 2>/dev/null || true
    gh label create "phase-3" --description "Phase 3: API Endpoints" --color "0e8a16" --force 2>/dev/null || true
    gh label create "phase-4" --description "Phase 4: Frontend Admin Dashboard" --color "1d76db" --force 2>/dev/null || true
    gh label create "phase-5" --description "Phase 5: Enforcement & Validation" --color "5319e7" --force 2>/dev/null || true
    gh label create "phase-6" --description "Phase 6: Testing & Documentation" --color "bfd4f2" --force 2>/dev/null || true
    gh label create "phase-7" --description "Phase 7: Migration & Deployment" --color "c2e0c6" --force 2>/dev/null || true
    
    echo -e "${GREEN}✓ Tags created${NC}"
}

# Function to create an issue
create_issue() {
    local issue_num=$1
    local title=$2
    local body=$3
    local labels=$4
    
    echo -e "${YELLOW}Creating issue #${issue_num}: ${title}${NC}"
    
    local issue_id=$(gh issue create \
        --title "${title}" \
        --body "${body}" \
        --label "${labels}" \
        --repo "${REPO}" \
        --json number \
        --jq '.number' 2>/dev/null || echo "")
    
    if [ -n "$issue_id" ]; then
        echo -e "${GREEN}✓ Issue #${issue_id} created${NC}"
        echo "$issue_id"
    else
        echo -e "${RED}✗ Failed to create issue #${issue_num}${NC}"
        echo ""
    fi
}

# Function to close an issue
close_issue() {
    local issue_num=$1
    local comment=$2
    
    echo -e "${YELLOW}Closing issue #${issue_num}...${NC}"
    
    if [ -n "$comment" ]; then
        gh issue comment "${issue_num}" --body "${comment}" --repo "${REPO}" 2>/dev/null || true
    fi
    
    gh issue close "${issue_num}" --repo "${REPO}" 2>/dev/null && \
        echo -e "${GREEN}✓ Issue #${issue_num} closed${NC}" || \
        echo -e "${RED}✗ Failed to close issue #${issue_num}${NC}"
}

# Function to get or create project board
get_or_create_project() {
    echo -e "${YELLOW}Setting up project board...${NC}"
    
    # Try to find existing project
    local project=$(gh project list --owner "$(gh repo view --json owner --jq '.owner.login')" --json number,title --jq '.[] | select(.title == "Capability Model Implementation") | .number' 2>/dev/null | head -n 1)
    
    if [ -n "$project" ]; then
        echo -e "${GREEN}✓ Found existing project board #${project}${NC}"
        PROJECT_NUMBER="$project"
    else
        echo -e "${YELLOW}Creating new project board...${NC}"
        # Note: Project creation via CLI may require GitHub API v2
        echo -e "${YELLOW}Note: Please create project board manually in GitHub UI${NC}"
        echo "Project name: Capability Model Implementation"
        echo "Columns: Backlog, In Progress, In Review, Done"
    fi
}

# Function to add issue to project
add_to_project() {
    local issue_num=$1
    local column=$2
    
    if [ -z "$PROJECT_NUMBER" ]; then
        echo -e "${YELLOW}Project board not set, skipping...${NC}"
        return
    fi
    
    echo -e "${YELLOW}Adding issue #${issue_num} to project...${NC}"
    # Note: Adding to project requires GitHub API v2 or manual action
    echo -e "${YELLOW}Note: Please add issue #${issue_num} to project board manually${NC}"
}

# Main menu
show_menu() {
    echo ""
    echo -e "${BLUE}Select an action:${NC}"
    echo "1) Create all tags"
    echo "2) Create Phase 1 issues (Database & Models)"
    echo "3) Create Phase 2 issues (Backend Core Logic)"
    echo "4) Create all issues (Phases 1-7)"
    echo "5) Close completed issues (Phase 1 & Phase 2 partial)"
    echo "6) Setup project board"
    echo "7) Exit"
    echo ""
    read -p "Enter choice [1-7]: " choice
    echo ""
    
    case $choice in
        1)
            create_tags
            ;;
        2)
            create_tags
            create_phase1_issues
            ;;
        3)
            create_tags
            create_phase2_issues
            ;;
        4)
            create_tags
            create_all_issues
            ;;
        5)
            close_completed_issues
            ;;
        6)
            get_or_create_project
            ;;
        7)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid choice${NC}"
            ;;
    esac
}

# Create Phase 1 issues
create_phase1_issues() {
    echo -e "${BLUE}Creating Phase 1 issues...${NC}"
    
    create_issue "001" \
        "[Phase 1] Create tenant_capabilities table" \
        "Create database migration and table for storing which capabilities are allowed for each tenant. This implements the \"System → Tenant\" layer of the capability model.

**Acceptance Criteria:**
- [ ] Migration file \`000018_create_tenant_capabilities.up.sql\` created
- [ ] Migration file \`000018_create_tenant_capabilities.down.sql\` created
- [ ] Table includes: tenant_id, capability_key, enabled, value (JSONB), configured_by, configured_at
- [ ] Primary key on (tenant_id, capability_key)
- [ ] Indexes created for tenant_id and capability_key
- [ ] Migration tested and verified" \
        "database,migration,p0,capability-model,phase-1"
    
    create_issue "002" \
        "[Phase 1] Create system_capabilities table" \
        "Create database migration and table for storing global system-level capabilities. This implements the \"System\" layer of the capability model.

**Acceptance Criteria:**
- [ ] Migration file \`000019_create_system_capabilities.up.sql\` created
- [ ] Migration file \`000019_create_system_capabilities.down.sql\` created
- [ ] Table includes: capability_key, enabled, default_value (JSONB), description, updated_by, updated_at
- [ ] Default capabilities inserted: mfa, totp, saml, oidc, oauth2, passwordless, ldap, max_token_ttl, allowed_grant_types, allowed_scope_namespaces, pkce_mandatory
- [ ] Migration tested and verified" \
        "database,migration,p0,capability-model,phase-1"
    
    create_issue "003" \
        "[Phase 1] Create tenant_feature_enablement table" \
        "Create database migration and table for storing which features tenants have actually enabled. This implements the \"Tenant\" layer of the capability model.

**Acceptance Criteria:**
- [ ] Migration file \`000020_create_tenant_feature_enablement.up.sql\` created
- [ ] Migration file \`000020_create_tenant_feature_enablement.down.sql\` created
- [ ] Table includes: tenant_id, feature_key, enabled, configuration (JSONB), enabled_by, enabled_at
- [ ] Primary key on (tenant_id, feature_key)
- [ ] Indexes created for tenant_id and feature_key
- [ ] Migration tested and verified" \
        "database,migration,p0,capability-model,phase-1"
    
    create_issue "004" \
        "[Phase 1] Create user_capability_state table" \
        "Create database migration and table for storing user-level capability enrollment state (e.g., TOTP secrets, MFA enrollment status).

**Acceptance Criteria:**
- [ ] Migration file \`000021_create_user_capability_state.up.sql\` created
- [ ] Migration file \`000021_create_user_capability_state.down.sql\` created
- [ ] Table includes: user_id, capability_key, enrolled, state_data (JSONB), enrolled_at, last_used_at
- [ ] Primary key on (user_id, capability_key)
- [ ] Indexes created for user_id and capability_key
- [ ] Migration tested and verified" \
        "database,migration,p0,capability-model,phase-1"
    
    create_issue "005" \
        "[Phase 1] Create Go models for capability tables" \
        "Create Go model structs and validation logic for system_capabilities, tenant_capabilities, tenant_feature_enablement, and user_capability_state tables.

**Acceptance Criteria:**
- [ ] Model file \`identity/models/system_capability.go\` created
- [ ] Model file \`identity/models/tenant_capability.go\` created
- [ ] Model file \`identity/models/tenant_feature_enablement.go\` created
- [ ] Model file \`identity/models/user_capability_state.go\` created
- [ ] All models include proper JSON tags and validation
- [ ] Models include helper methods (IsEnabled, GetValue, etc.)
- [ ] Unit tests for models" \
        "backend,models,p0,capability-model,phase-1"
}

# Create Phase 2 issues
create_phase2_issues() {
    echo -e "${BLUE}Creating Phase 2 issues...${NC}"
    
    create_issue "006" \
        "[Phase 2] Implement capability evaluation service" \
        "Create the core capability service that evaluates capabilities across System → Tenant → User layers. This is the heart of the capability model.

**Acceptance Criteria:**
- [ ] Service file \`identity/capability/service.go\` created
- [ ] Interface \`CapabilityService\` defined with all required methods
- [ ] Implementation handles System level checks
- [ ] Implementation handles System→Tenant level checks
- [ ] Implementation handles Tenant level checks
- [ ] Implementation handles User level checks
- [ ] \`EvaluateCapability\` method combines all levels correctly
- [ ] Comprehensive unit tests (90%+ coverage)
- [ ] Integration tests with database" \
        "backend,service,p0,capability-model,phase-2"
    
    create_issue "007" \
        "[Phase 2] Implement capability repositories" \
        "Create repository interfaces and PostgreSQL implementations for all capability-related tables.

**Acceptance Criteria:**
- [ ] Interface files created in \`storage/interfaces/\`
- [ ] Implementation files created in \`storage/postgres/\`
- [ ] All CRUD operations implemented
- [ ] Proper error handling
- [ ] Unit tests for repositories
- [ ] Integration tests with database" \
        "backend,repository,p0,capability-model,phase-2"
    
    create_issue "008" \
        "[Phase 2] Integrate capability checks in auth flow" \
        "Update authentication flow to check capabilities before allowing login, MFA, and token issuance.

**Acceptance Criteria:**
- [ ] \`auth/login/service.go\` checks if password auth is allowed
- [ ] \`auth/login/service.go\` checks if MFA is required and allowed
- [ ] \`auth/mfa/mfa.go\` enforces MFA based on capabilities
- [ ] \`auth/token/token.go\` validates scopes against allowed namespaces
- [ ] Error messages are clear and actionable
- [ ] Integration tests for auth flow with capabilities
- [ ] E2E tests for login with capability restrictions" \
        "backend,authentication,p0,capability-model,phase-2,security"
    
    create_issue "009" \
        "[Phase 2] Integrate capability checks in OAuth flow" \
        "Update OAuth/OIDC flow to validate grant types, scopes, and PKCE requirements based on capabilities.

**Acceptance Criteria:**
- [ ] \`auth/hydra/hydra.go\` validates grant types against allowed list
- [ ] Scope validation checks against allowed scope namespaces
- [ ] PKCE enforcement based on system capability
- [ ] OAuth client creation checks if OIDC/OAuth2 is allowed
- [ ] Integration tests for OAuth flow with capabilities
- [ ] E2E tests for OAuth with capability restrictions" \
        "backend,oauth,p0,capability-model,phase-2,security"
}

# Create all issues (placeholder - would need all 30 issues)
create_all_issues() {
    echo -e "${BLUE}Creating all issues...${NC}"
    create_phase1_issues
    create_phase2_issues
    echo -e "${YELLOW}Note: Issues 010-030 would be created here. See docs/planning/GITHUB_ISSUES.md for full list.${NC}"
}

# Close completed issues
close_completed_issues() {
    echo -e "${BLUE}Closing completed issues...${NC}"
    
    # Phase 1 - All completed
    close_issue "001" "✅ Completed: Migration 000018 created with all required fields and indexes"
    close_issue "002" "✅ Completed: Migration 000019 created with default capabilities inserted"
    close_issue "003" "✅ Completed: Migration 000020 created with all required fields"
    close_issue "004" "✅ Completed: Migration 000021 created with all required fields"
    close_issue "005" "✅ Completed: All 4 Go models created with helper methods"
    
    # Phase 2 - Partial completion
    close_issue "006" "✅ Completed: Capability service implemented with full three-layer evaluation"
    close_issue "007" "✅ Completed: All 4 repository interfaces and implementations created"
    
    echo -e "${GREEN}✓ Completed issues closed${NC}"
}

# Run main menu
if [ "$1" == "--auto" ]; then
    # Auto mode: create tags and close completed issues
    create_tags
    close_completed_issues
else
    # Interactive mode
    while true; do
        show_menu
    done
fi

