# GitHub Management Guide for Capability Model Implementation

This guide explains how to manage GitHub issues, tags, and project board for the Capability Model implementation.

---

## 🚀 Quick Start

### Automated Management

Use the provided script to manage issues:

```bash
# Interactive mode (recommended)
./scripts/manage-github-capability-issues.sh

# Auto mode (creates tags and closes completed issues)
./scripts/manage-github-capability-issues.sh --auto
```

---

## 📋 Manual Steps

### 1. Create GitHub Tags

All tags are defined in `docs/planning/GITHUB_TAGS.md`. Use the script or create manually:

```bash
# Priority tags
gh label create "p0" --description "Critical priority" --color "d73a4a"
gh label create "p1" --description "Important priority" --color "fbca04"
gh label create "p2" --description "Nice to have" --color "0e8a16"

# Component tags
gh label create "backend" --description "Backend code changes" --color "1d76db"
gh label create "frontend" --description "Frontend code changes" --color "bfd4f2"
gh label create "database" --description "Database changes" --color "5319e7"
# ... (see GITHUB_TAGS.md for full list)
```

### 2. Create Issues

Issues are documented in `docs/planning/GITHUB_ISSUES.md`. Create them using:

**Option A: GitHub CLI**
```bash
gh issue create \
  --title "[Phase 1] Create tenant_capabilities table" \
  --body "$(cat issue-001-body.md)" \
  --label "database,migration,p0,capability-model,phase-1"
```

**Option B: GitHub Web UI**
1. Go to Issues → New Issue
2. Use templates from `docs/planning/GITHUB_ISSUES.md`
3. Add appropriate labels

**Option C: Use the management script**
```bash
./scripts/manage-github-capability-issues.sh
# Select option 2 for Phase 1 issues
```

### 3. Create Project Board

1. Go to your repository on GitHub
2. Click "Projects" → "New project"
3. Name: "Capability Model Implementation"
4. Create columns:
   - **Backlog** - Issues not yet started
   - **In Progress** - Issues actively being worked on
   - **In Review** - Issues completed, awaiting review
   - **Done** - Issues completed and verified

### 4. Add Issues to Project Board

**Option A: Via GitHub Web UI**
1. Open the issue
2. Click "Projects" in the sidebar
3. Select "Capability Model Implementation"
4. Choose the appropriate column

**Option B: Via GitHub CLI** (requires API v2)
```bash
gh api graphql -f query='
  mutation {
    addProjectV2ItemById(input: {
      projectId: "PROJECT_ID"
      contentId: "ISSUE_ID"
    }) {
      item {
        id
      }
    }
  }
'
```

### 5. Close Completed Issues

**Option A: Use the script**
```bash
./scripts/manage-github-capability-issues.sh
# Select option 5
```

**Option B: Manual**
```bash
gh issue close 001 --comment "✅ Completed: Migration 000018 created"
gh issue close 002 --comment "✅ Completed: Migration 000019 created"
# ... etc
```

---

## 📊 Current Status

### Phase 1: Database & Models ✅ (100% Complete)
- ✅ Issue #001: Create tenant_capabilities table
- ✅ Issue #002: Create system_capabilities table
- ✅ Issue #003: Create tenant_feature_enablement table
- ✅ Issue #004: Create user_capability_state table
- ✅ Issue #005: Create Go models for capability tables

### Phase 2: Backend Core Logic 🟡 (50% Complete)
- ✅ Issue #006: Implement capability evaluation service
- ✅ Issue #007: Implement capability repositories
- ⏳ Issue #008: Integrate capability checks in auth flow (In Progress)
- ⏳ Issue #009: Integrate capability checks in OAuth flow (Pending)

---

## 🔄 Workflow

### When Starting Work on an Issue

1. **Move issue to "In Progress"** on project board
2. **Assign yourself** to the issue
3. **Create a branch**: `git checkout -b issue-001-create-tenant-capabilities-table`
4. **Work on the issue**
5. **Update status** in `docs/status/CAPABILITY_MODEL_STATUS.md`

### When Completing an Issue

1. **Create PR** with issue number in title: `[#001] Create tenant_capabilities table`
2. **Link PR to issue**: Add "Closes #001" in PR description
3. **Move issue to "In Review"** on project board
4. **After merge**: Issue auto-closes, move to "Done" on board

### When Closing Issues Manually

```bash
# Close with comment
gh issue close 001 --comment "✅ Completed: [description of what was done]"

# Or use the script
./scripts/manage-github-capability-issues.sh
# Select option 5
```

---

## 📝 Issue Template

When creating issues manually, use this template:

```markdown
## Description
[Brief description of what needs to be done]

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Dependencies
- Depends on #XXX
- Related to #YYY

## Notes
[Any additional notes or context]
```

---

## 🏷️ Tag Usage

### Always Include
- `capability-model` - All capability model issues
- One phase tag (`phase-1` through `phase-7`)
- One priority tag (`p0`, `p1`, or `p2`)

### Component Tags (as applicable)
- `backend` - Backend changes
- `frontend` - Frontend changes
- `database` - Database/migration changes
- `api` - API endpoint changes
- `testing` - Test-related work
- `documentation` - Documentation updates

### Feature Tags (as applicable)
- `system` - System-level features
- `tenant` - Tenant-level features
- `user` - User-level features
- `mfa` - MFA/TOTP features
- `oauth` - OAuth2/OIDC features
- `security` - Security-related

---

## 🔗 Related Documents

- [Implementation Plan](CAPABILITY_MODEL_IMPLEMENTATION_PLAN.md)
- [GitHub Issues](GITHUB_ISSUES.md) - All 30 issues with details
- [GitHub Tags](GITHUB_TAGS.md) - Complete tag structure
- [Status Tracking](../status/CAPABILITY_MODEL_STATUS.md) - Current progress

---

## 🛠️ Troubleshooting

### Issue: "gh: command not found"
**Solution**: Install GitHub CLI from https://cli.github.com/

### Issue: "Not authenticated"
**Solution**: Run `gh auth login`

### Issue: "Repository not found"
**Solution**: Check repository name in script (`REPO` variable)

### Issue: "Cannot create project via CLI"
**Solution**: Create project board manually in GitHub UI, then use script for issues

---

**Last Updated**: 2025-01-27

