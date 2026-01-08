# GitHub Quick Reference - Capability Model

Quick commands for managing GitHub issues, tags, and project board.

---

## 🏷️ Create All Tags

```bash
./scripts/manage-github-capability-issues.sh
# Select option 1
```

Or manually:
```bash
gh label create "p0" --description "Critical priority" --color "d73a4a" --force
gh label create "p1" --description "Important priority" --color "fbca04" --force
gh label create "capability-model" --description "Core capability model" --color "b60205" --force
# ... (see script for full list)
```

---

## 📝 Create Issues

### Phase 1 Issues (All Completed ✅)
```bash
./scripts/manage-github-capability-issues.sh
# Select option 2
```

### Phase 2 Issues (Partial ✅)
```bash
./scripts/manage-github-capability-issues.sh
# Select option 3
```

### All Issues
```bash
./scripts/manage-github-capability-issues.sh
# Select option 4
```

---

## ✅ Close Completed Issues

```bash
./scripts/manage-github-capability-issues.sh
# Select option 5
```

Or manually:
```bash
gh issue close 001 --comment "✅ Completed: Migration 000018 created"
gh issue close 002 --comment "✅ Completed: Migration 000019 created"
gh issue close 003 --comment "✅ Completed: Migration 000020 created"
gh issue close 004 --comment "✅ Completed: Migration 000021 created"
gh issue close 005 --comment "✅ Completed: All 4 Go models created"
gh issue close 006 --comment "✅ Completed: Capability service implemented"
gh issue close 007 --comment "✅ Completed: All 4 repositories created"
```

---

## 📊 View Issues

```bash
# List all capability model issues
gh issue list --label "capability-model"

# List Phase 1 issues
gh issue list --label "phase-1"

# List open issues
gh issue list --state open --label "capability-model"

# List closed issues
gh issue list --state closed --label "capability-model"
```

---

## 🔄 Update Issue Status

```bash
# Add comment
gh issue comment 008 --body "Working on integrating capability checks in login service"

# Add label
gh issue edit 008 --add-label "in-progress"

# Assign to user
gh issue edit 008 --add-assignee "@me"
```

---

## 📋 Project Board

### Create Project (Manual)
1. Go to repository → Projects → New project
2. Name: "Capability Model Implementation"
3. Add columns: Backlog, In Progress, In Review, Done

### Add Issue to Project (Manual)
1. Open issue
2. Click "Projects" in sidebar
3. Select project and column

---

## 🎯 Current Status Summary

**Completed Issues**: 7/30 (23%)
- Phase 1: 5/5 (100%) ✅
- Phase 2: 2/4 (50%) 🟡

**Next Steps**:
1. Close completed issues (#001-#007)
2. Continue with #008 (Integrate capability checks in auth flow)
3. Then #009 (Integrate capability checks in OAuth flow)

---

## 📚 Full Documentation

- [GitHub Management Guide](GITHUB_MANAGEMENT.md) - Detailed guide
- [GitHub Issues](GITHUB_ISSUES.md) - All 30 issues
- [GitHub Tags](GITHUB_TAGS.md) - Tag structure
- [Status Tracking](../status/CAPABILITY_MODEL_STATUS.md) - Current progress

---

**Last Updated**: 2025-01-27

