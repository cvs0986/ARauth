# ARauth IAM Control Plane - Guardrails Enforcement

**Document Type**: Compliance & Quality Standards  
**Status**: MANDATORY - Non-Negotiable  
**Last Updated**: 2026-01-11

---

## 🔒 The 7 Guardrails (MANDATORY)

### Guardrail 1: Backend Is Law

**Rule**: The admin console MUST NOT invent security logic, permissions, or behavior.

**Enforcement**:
- ✅ **DO**: Read permissions from JWT claims via PrincipalContext
- ✅ **DO**: Call backend APIs to fetch data
- ✅ **DO**: Display backend-provided security states
- ❌ **DON'T**: Compute permissions in frontend
- ❌ **DON'T**: Assume security states
- ❌ **DON'T**: Implement authorization logic

**Stop Conditions**:
- If permission behavior is unclear → **STOP and ASK**
- If API response structure is ambiguous → **STOP and ASK**
- If security policy is not documented → **STOP and ASK**

**Code Review Checklist**:
- [ ] No `if (user.role === 'admin')` logic in UI
- [ ] All permissions checked via `hasPermission()` or `hasSystemPermission()`
- [ ] No hardcoded role names in components
- [ ] All authority decisions delegated to backend

---

### Guardrail 2: No UI Security Semantics

**Rule**: UI may visualize security state, never decide it.

**Forbidden**:
- ❌ MFA enforcement logic in UI
- ❌ Session trust assumptions
- ❌ Token validity assumptions
- ❌ "Derived" security states (e.g., `isSecure = mfaEnabled && passwordStrong`)

**Allowed**:
- ✅ Display MFA status from backend
- ✅ Show session expiry from token claims
- ✅ Visualize security posture from backend metrics

**Examples**:

**❌ WRONG**:
```typescript
// DON'T compute security state in UI
const isSecure = user.mfaEnabled && user.passwordStrength === 'strong';
```

**✅ CORRECT**:
```typescript
// DO fetch security state from backend
const { data: securityStatus } = useQuery({
  queryKey: ['security-status', userId],
  queryFn: () => api.getSecurityStatus(userId),
});
```

**Code Review Checklist**:
- [ ] No MFA validation logic in frontend
- [ ] No session timeout calculations (use backend expiry)
- [ ] No token parsing except via PrincipalContext
- [ ] All security metrics fetched from backend

---

### Guardrail 3: Feature Discovery Is Locked

**Rule**: The feature discovery section is authoritative. No silent additions.

**Process**:
1. If new backend capability discovered → Document in feature inventory
2. Create GitHub issue for UI implementation
3. Get approval before starting UI work
4. Update `iam_control_plane_complete.md`

**Locked Features** (from discovery):
- ✅ Authentication (password, MFA, tokens)
- ✅ Authorization (roles, permissions, capabilities)
- ✅ OAuth2/OIDC (scopes, clients - UI missing)
- ✅ SAML (federation - UI missing)
- ✅ SCIM (provisioning - UI missing)
- ✅ Webhooks
- ✅ Impersonation
- ✅ Identity Linking
- ✅ Audit Logs

**If You Discover New Features**:
1. **STOP** implementation
2. Document in `FEATURE_DISCOVERY_ADDENDUM.md`
3. Create GitHub issue
4. Get approval
5. Update implementation plan
6. Resume work

**Code Review Checklist**:
- [ ] All UI features map to documented backend features
- [ ] No "experimental" features without approval
- [ ] Feature discovery document is up-to-date

---

### Guardrail 4: Data Gaps Must Be Explicit

**Rule**: If dashboard metric requires non-existent backend aggregation, stub it.

**Process**:
1. Identify missing backend API
2. Add "Coming Soon" placeholder in UI
3. Create GitHub issue for backend work
4. Document in `DATA_GAPS.md`

**Examples**:

**❌ WRONG**:
```typescript
// DON'T fake aggregation in frontend
const totalUsers = tenants.reduce((sum, t) => sum + t.users.length, 0);
```

**✅ CORRECT**:
```typescript
// DO stub with explicit gap
<MetricCard
  title="Total Users"
  value="Coming Soon"
  tooltip="Requires backend aggregation API (Issue #123)"
/>
```

**Known Data Gaps**:
- ⚠️ Cross-tenant user aggregation (SYSTEM dashboard)
- ⚠️ MFA adoption rate calculation (needs backend endpoint)
- ⚠️ Security posture scoring (needs backend logic)
- ⚠️ Tenant health metrics (needs backend computation)

**Code Review Checklist**:
- [ ] No client-side aggregation of server data
- [ ] All "Coming Soon" placeholders have GitHub issues
- [ ] Data gaps documented in `DATA_GAPS.md`

---

### Guardrail 5: Vertical Slices Only

**Rule**: Execution order is non-negotiable.

**Mandatory Order**:
1. ✅ PrincipalContext (foundation)
2. ✅ ConsoleMode (mode switching)
3. ✅ Header + TenantSelector (navigation chrome)
4. ✅ Sidebar (navigation tree)
5. ⏳ Dashboards (read-only views)
6. ⏳ Settings (configuration)
7. ⏳ Protocols (OAuth2, SCIM, SAML)
8. ⏳ Workflows (guided multi-step)
9. ⏳ Polish & tests

**No Skipping Ahead**:
- ❌ Can't build OAuth2 clients page before Sidebar is done
- ❌ Can't build workflows before Settings is done
- ❌ Can't polish before core features are complete

**Why This Matters**:
- Each layer depends on previous layers
- Skipping creates technical debt
- Quality compounds with each layer

**Code Review Checklist**:
- [ ] Current phase is complete before starting next
- [ ] No orphaned components from future phases
- [ ] Dependencies are satisfied

---

### Guardrail 6: UI Quality Bar Is Enforced

**Rule**: This is a control plane, not an admin toy.

**Quality Standards**:

**❌ REJECT if**:
- Looks like CRUD (generic table with edit/delete buttons)
- Feels cluttered (too much info, poor hierarchy)
- Hides authority context (user doesn't know their mode)
- Has disabled buttons (hide actions user can't perform)
- Uses generic error messages ("Error occurred")
- Lacks loading states (instant transitions)
- Has no empty states (blank screens)

**✅ ACCEPT if**:
- Feels operator-grade (calm, confident, purposeful)
- Shows clear authority context (mode, tenant, permissions)
- Uses permission-gated visibility (not disabled buttons)
- Has specific error messages ("MFA required for this action")
- Has skeleton loaders (smooth transitions)
- Has guided empty states ("Create your first tenant")

**Examples**:

**❌ WRONG**:
```typescript
<Button disabled={!canCreateUser}>Create User</Button>
```

**✅ CORRECT**:
```typescript
{canCreateUser && <Button>Create User</Button>}
```

**Code Review Checklist**:
- [ ] No disabled buttons (use conditional rendering)
- [ ] Authority context always visible (mode badge, tenant name)
- [ ] Skeleton loaders on all async operations
- [ ] Empty states with actionable guidance
- [ ] Error messages are specific and helpful
- [ ] UI feels calm and professional

---

### Guardrail 7: GitHub Hygiene Is Mandatory

**Rule**: If GitHub stops being "alive", quality will decay.

**Requirements**:

**Issues**:
- ✅ One issue per phase/feature
- ✅ Clear acceptance criteria
- ✅ Labels: `phase-1`, `frontend`, `control-plane`
- ✅ Linked to project board

**Branches**:
- ✅ Feature branches: `feature/principal-context`
- ✅ One branch per issue
- ✅ Branch from `main`

**Commits**:
- ✅ Frequent commits (daily minimum)
- ✅ Descriptive messages: `feat: implement PrincipalContext`
- ✅ Atomic commits (one logical change)

**Pull Requests**:
- ✅ One PR per feature
- ✅ Description links to issue
- ✅ Screenshots for UI changes
- ✅ Tests pass before merge

**Documentation**:
- ✅ Update docs alongside code
- ✅ Update `STATUS.md` after each phase
- ✅ Keep `task.md` current

**Kanban**:
- ✅ Columns: Backlog, In Progress, Review, Done
- ✅ Move cards as work progresses
- ✅ Weekly review

**Code Review Checklist**:
- [ ] GitHub issue exists for this work
- [ ] Branch name matches issue
- [ ] Commits are descriptive
- [ ] PR has screenshots (if UI)
- [ ] Docs are updated
- [ ] Kanban is current

---

## 🚨 Stop Conditions (When to STOP and ASK)

**STOP immediately if**:
1. Backend API behavior is unclear
2. Permission model is ambiguous
3. Security policy is not documented
4. New feature discovered that's not in inventory
5. Data aggregation requires backend work
6. UI quality doesn't meet standards
7. GitHub process is being skipped

**How to ASK**:
1. Create GitHub issue with `question` label
2. Tag in PR comments
3. Block PR until clarified
4. Document decision in ADR (Architecture Decision Record)

---

## ✅ Compliance Verification

**Before Each PR**:
- [ ] Guardrail 1: No invented security logic
- [ ] Guardrail 2: No UI security semantics
- [ ] Guardrail 3: Feature is in discovery doc
- [ ] Guardrail 4: Data gaps are stubbed
- [ ] Guardrail 5: Vertical slice order followed
- [ ] Guardrail 6: UI quality standards met
- [ ] Guardrail 7: GitHub hygiene maintained

**Before Each Phase**:
- [ ] All issues created
- [ ] Kanban board updated
- [ ] Documentation current
- [ ] Previous phase complete

**Before Production**:
- [ ] All guardrails verified
- [ ] Security review passed
- [ ] Performance benchmarks met
- [ ] Accessibility audit passed (WCAG 2.1 AA)

---

## 📋 Guardrails Checklist (Print & Post)

```
┌─────────────────────────────────────────────┐
│  ARauth IAM Control Plane - Guardrails     │
├─────────────────────────────────────────────┤
│  □ Backend Is Law (no invented logic)      │
│  □ No UI Security Semantics                │
│  □ Feature Discovery Is Locked             │
│  □ Data Gaps Are Explicit                  │
│  □ Vertical Slices Only                    │
│  □ UI Quality Bar Enforced                 │
│  □ GitHub Hygiene Maintained               │
└─────────────────────────────────────────────┘
```

---

**These guardrails are non-negotiable. Violations require rework.**
