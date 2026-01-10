#!/bin/bash

# Script to create GitHub issues for Phase 1 implementation
# Usage: ./scripts/create-github-issues.sh

set -e

REPO="cvs0986/ARauth"
BASE_URL="https://api.github.com/repos/${REPO}/issues"

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "GitHub CLI (gh) is not installed. Please install it first."
    echo "Visit: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "Not authenticated with GitHub. Please run: gh auth login"
    exit 1
fi

echo "Creating GitHub issues for Phase 1 implementation..."

# Issue 1: Audit Events System
gh issue create \
  --title "Implement Structured Audit Events System" \
  --body "## Overview
Implement a structured audit events system to track all important actions in the IAM system.

## Requirements
- [x] Database schema for audit_events table
- [x] Models (AuditEvent, AuditActor, AuditTarget)
- [x] Repository interface and implementation
- [x] Service layer with helper methods
- [x] API handlers (QueryEvents, GetEvent)
- [x] Routes (tenant-scoped and system-wide)
- [ ] Integration with all handlers
- [ ] Testing
- [ ] Documentation

## Implementation Plan
See: \`docs/implementation/FUTURE_FEATURES_IMPLEMENTATION_PLAN.md\` section 1

## Status
🚧 In Progress (60% complete)

## Related
- Part of Phase 1: Critical Missing Features
- Estimated: 3-5 days
- Priority: HIGH" \
  --label "enhancement,phase-1,high-priority,audit" \
  --assignee "@me"

# Issue 2: Federation (OIDC/SAML)
gh issue create \
  --title "Implement Federation (OIDC/SAML Login)" \
  --body "## Overview
Implement external identity provider integration for OIDC and SAML federation.

## Requirements
- [ ] OIDC provider configuration
- [ ] SAML IdP configuration
- [ ] OIDC login flow
- [ ] SAML SSO flow
- [ ] Identity provider management API
- [ ] Token exchange
- [ ] Attribute mapping

## Implementation Plan
See: \`docs/implementation/FUTURE_FEATURES_IMPLEMENTATION_PLAN.md\` section 3

## Status
⏸️ Pending

## Related
- Part of Phase 1: Critical Missing Features
- Estimated: 10-15 days
- Priority: HIGH" \
  --label "enhancement,phase-1,high-priority,federation" \
  --assignee "@me"

# Issue 3: Event Hooks / Webhooks
gh issue create \
  --title "Implement Event Hooks / Webhooks System" \
  --body "## Overview
Implement a webhook system to notify external systems of important events.

## Requirements
- [ ] Webhook configuration API
- [ ] Event subscriptions
- [ ] Retry logic with exponential backoff
- [ ] Webhook secret signing
- [ ] Delivery status tracking
- [ ] Webhook dispatcher (async)

## Implementation Plan
See: \`docs/implementation/FUTURE_FEATURES_IMPLEMENTATION_PLAN.md\` section 2

## Status
⏸️ Pending

## Related
- Part of Phase 1: Critical Missing Features
- Estimated: 5-7 days
- Priority: MEDIUM" \
  --label "enhancement,phase-1,medium-priority,webhooks" \
  --assignee "@me"

# Issue 4: Identity Linking
gh issue create \
  --title "Implement Identity Linking" \
  --body "## Overview
Allow users to have multiple identities (password + SAML + OIDC) linked to one account.

## Requirements
- [ ] Link/unlink identities
- [ ] Primary identity designation
- [ ] Identity verification
- [ ] Login flow with multiple identities

## Implementation Plan
See: \`docs/implementation/FUTURE_FEATURES_IMPLEMENTATION_PLAN.md\` section 4

## Status
⏸️ Pending

## Related
- Part of Phase 1: Critical Missing Features
- Estimated: 3-4 days
- Priority: MEDIUM" \
  --label "enhancement,phase-1,medium-priority,identity-linking" \
  --assignee "@me"

echo ""
echo "✅ GitHub issues created successfully!"
echo ""
echo "View issues: https://github.com/${REPO}/issues"

