#!/bin/bash

# Script to create GitHub issues for frontend development
# Usage: ./scripts/create-frontend-issues.sh

set -e

echo "🚀 Creating Frontend Development Issues..."

# Issue 11: Tenant Management UI
gh issue create \
  --title "Frontend: Admin Dashboard - Tenant Management UI" \
  --body "Build tenant management UI with CRUD operations

## Tasks
- [ ] List tenants page with table
- [ ] Create tenant form
- [ ] Edit tenant form
- [ ] Delete tenant with confirmation
- [ ] Tenant search and filtering
- [ ] Pagination

## Acceptance Criteria
- All tenant CRUD operations work via UI
- Forms have proper validation
- Error handling implemented
- Responsive design" \
  || echo "Issue #11 creation failed"

# Issue 12: User Management UI
gh issue create \
  --title "Frontend: Admin Dashboard - User Management UI" \
  --body "Build user management UI with CRUD operations

## Tasks
- [ ] List users page with table
- [ ] Create user form with validation
- [ ] Edit user form
- [ ] Delete user with confirmation
- [ ] User search and filtering
- [ ] Role assignment UI

## Acceptance Criteria
- All user CRUD operations work via UI
- Password validation
- Role assignment works
- Responsive design" \
  || echo "Issue #12 creation failed"

# Issue 13: Role Management UI
gh issue create \
  --title "Frontend: Admin Dashboard - Role Management UI" \
  --body "Build role management UI

## Tasks
- [ ] List roles page
- [ ] Create role form
- [ ] Edit role form
- [ ] Permission assignment UI
- [ ] Role details view

## Acceptance Criteria
- All role CRUD operations work via UI
- Permission assignment works
- Role details display correctly" \
  || echo "Issue #13 creation failed"

# Issue 14: Permission Management UI
gh issue create \
  --title "Frontend: Admin Dashboard - Permission Management UI" \
  --body "Build permission management UI

## Tasks
- [ ] List permissions page
- [ ] Create permission form
- [ ] Edit permission form
- [ ] Permission tree view

## Acceptance Criteria
- All permission CRUD operations work via UI
- Permission tree displays correctly" \
  || echo "Issue #14 creation failed"

# Issue 15: E2E Testing App - Authentication
gh issue create \
  --title "Frontend: E2E Testing App - Authentication Flow" \
  --body "Build authentication flow for testing

## Tasks
- [ ] Registration page
- [ ] Login page
- [ ] Logout functionality
- [ ] Token management
- [ ] Error handling

## Acceptance Criteria
- Complete registration flow works
- Login flow works
- Token storage and refresh works" \
  || echo "Issue #15 creation failed"

echo "✅ Frontend issues creation complete!"
echo "View issues: gh issue list"

