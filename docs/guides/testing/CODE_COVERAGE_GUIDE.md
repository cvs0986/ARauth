# 📊 Code Coverage on GitHub

## Where to View Code Coverage

### 1. **Pull Requests** (Primary Location)

Once Codecov is set up, you'll see coverage reports directly in Pull Requests:

- **Coverage badge** at the top of the PR
- **File-by-file coverage** in the "Files changed" tab
- **Coverage diff** showing how coverage changed
- **Comments** with coverage summary

**Location**: https://github.com/cvs0986/ARauth/pulls

### 2. **GitHub Actions** (Workflow Runs)

View coverage in the Actions tab:

1. Go to: https://github.com/cvs0986/ARauth/actions
2. Click on a workflow run
3. Look for the "Upload coverage" step
4. Coverage file is uploaded (but may need Codecov to view nicely)

### 3. **Codecov Dashboard** (Best Option)

If Codecov is connected, you'll have a dedicated dashboard:

- **Repository Dashboard**: https://codecov.io/gh/cvs0986/ARauth
- **Coverage trends** over time
- **File-by-file coverage** breakdown
- **Coverage badges** for README

## 🔧 Setting Up Codecov

Your CI workflow is already configured to upload coverage! You just need to connect Codecov:

### Step 1: Sign Up for Codecov

1. Go to: https://codecov.io
2. Sign in with your GitHub account
3. Authorize Codecov to access your repositories

### Step 2: Add Your Repository

1. In Codecov dashboard, click "Add a repository"
2. Find `cvs0986/ARauth`
3. Click "Set up repository"
4. Codecov will automatically detect your uploads

### Step 3: Get Coverage Token (Optional)

If you need a token for private repos or advanced features:

1. Go to repository settings in Codecov
2. Copy the upload token
3. Add to GitHub Secrets (if needed)

### Step 4: Verify Setup

After the next CI run:
- Coverage will appear in PRs automatically
- Dashboard will show coverage metrics
- Badge will be available for README

## 📈 Current Coverage Status

Based on your `TESTING_STATUS.md`:
- **Overall Coverage**: 80%
- **Total Tests**: 134+ tests
- **Unit Tests**: 114+ tests
- **Integration Tests**: 20 tests

## 🎯 Coverage Goals

From your testing strategy:
- **Overall**: > 80% ✅ (Currently at 80%)
- **Business Logic**: > 90%
- **API Handlers**: > 80%
- **Repositories**: > 85%

## 🔍 Viewing Coverage Locally

You can also view coverage locally:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# View summary
go tool cover -func=coverage.out
```

## 📊 Coverage Badge for README

Once Codecov is set up, you can add a badge to your README:

```markdown
[![codecov](https://codecov.io/gh/cvs0986/ARauth/branch/main/graph/badge.svg)](https://codecov.io/gh/cvs0986/ARauth)
```

## 🚀 Quick Check

To see if Codecov is already connected:

1. **Check PRs**: Look at any open PR - do you see coverage comments?
2. **Check Actions**: Run `gh run list` and check if coverage upload succeeds
3. **Visit Codecov**: Go to https://codecov.io/gh/cvs0986/ARauth

## 📝 Your CI Configuration

Your `.github/workflows/ci.yml` already has:

```yaml
- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out
    fail_ci_if_error: false
```

This means coverage is being generated and uploaded! You just need Codecov connected to view it.

---

**Next Steps**: 
1. Visit https://codecov.io and connect your repository
2. Wait for the next CI run
3. Coverage will appear in PRs automatically!

# Code Coverage
