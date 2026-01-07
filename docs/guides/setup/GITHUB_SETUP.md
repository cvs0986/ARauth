# 🔗 GitHub Integration Setup Guide

This guide will help you connect your GitHub account to enable seamless development workflows with issue tracking, task management, and automated CI/CD.

## 📋 Prerequisites

1. **GitHub Account**: You need a GitHub account
2. **GitHub Repository**: Create a repository on GitHub (or use an existing one)
3. **Git**: Already initialized in this project ✅

## 🚀 Step-by-Step Setup

### Step 1: Create GitHub Repository

1. Go to [GitHub](https://github.com) and sign in
2. Click the **"+"** icon in the top right → **"New repository"**
3. Name it: `nuage-identity` (or your preferred name)
4. Choose **Public** or **Private**
5. **DO NOT** initialize with README, .gitignore, or license (we already have files)
6. Click **"Create repository"**

### Step 2: Connect Local Repository to GitHub

Run these commands in your terminal:

```bash
cd /home/eshwar/Documents/Veer/nuage-indentity

# Add remote (replace YOUR_USERNAME with your GitHub username)
git remote add origin https://github.com/YOUR_USERNAME/nuage-identity.git

# Or if using SSH (recommended for better security):
# git remote add origin git@github.com:YOUR_USERNAME/nuage-identity.git

# Verify remote is added
git remote -v
```

### Step 3: Initial Commit and Push

```bash
# Stage all files
git add .

# Create initial commit
git commit -m "Initial commit: Headless IAM platform with ORY Hydra"

# Push to GitHub (first time)
git branch -M main  # Rename branch to main if needed
git push -u origin main
```

### Step 4: Connect GitHub to Cursor

#### Option A: Using Cursor's Built-in GitHub Integration

1. Open Cursor Settings:
   - **Linux**: `Ctrl + ,` or `File → Preferences → Settings`
   - Look for **"GitHub"** in settings
   - Sign in with your GitHub account

2. Enable GitHub Features:
   - GitHub Copilot (if available)
   - GitHub Issues integration
   - Pull Request integration

#### Option B: Using GitHub CLI (Recommended)

Install GitHub CLI for command-line access:

```bash
# On Fedora/RHEL
sudo dnf install gh

# Or download from: https://cli.github.com/
```

Authenticate:

```bash
gh auth login
# Follow the prompts to authenticate
```

Verify connection:

```bash
gh auth status
```

### Step 5: Set Up GitHub Personal Access Token (For API Access)

If you want programmatic access to create issues, tasks, etc.:

1. Go to GitHub → **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Click **"Generate new token (classic)"**
3. Give it a name: `Cursor IAM Project`
4. Select scopes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `workflow` (Update GitHub Action workflows)
   - ✅ `write:packages` (if needed)
5. Click **"Generate token"**
6. **Copy the token immediately** (you won't see it again!)

Store it securely:

```bash
# Option 1: Environment variable (add to ~/.bashrc or ~/.zshrc)
export GITHUB_TOKEN="your_token_here"

# Option 2: GitHub CLI
gh auth login --with-token < your_token_file.txt
```

## 🎯 Using GitHub Features

### Creating Issues

#### Via GitHub Web Interface:
1. Go to your repository on GitHub
2. Click **"Issues"** tab
3. Click **"New issue"**
4. Choose a template (Bug Report, Feature Request, or Task)
5. Fill in the details and submit

#### Via GitHub CLI:
```bash
# Create a bug report
gh issue create --title "Bug: Login fails with special characters" \
  --body-file .github/ISSUE_TEMPLATE/bug_report.md \
  --label bug

# Create a feature request
gh issue create --title "Feature: Add SAML support" \
  --body-file .github/ISSUE_TEMPLATE/feature_request.md \
  --label enhancement

# Create a task
gh issue create --title "Task: Implement MFA service" \
  --body-file .github/ISSUE_TEMPLATE/task.md \
  --label task
```

#### Via Cursor (if integrated):
- Use the GitHub extension in Cursor
- Right-click in the editor → "Create GitHub Issue"
- Or use the command palette: `Ctrl+Shift+P` → "GitHub: Create Issue"

### Managing Tasks with GitHub Projects

1. Go to your repository → **"Projects"** tab
2. Click **"New project"**
3. Choose **"Board"** template
4. Add columns:
   - 📋 Backlog
   - 🔄 In Progress
   - 👀 Review
   - ✅ Done
5. Link issues to the project

### Using GitHub Actions (CI/CD)

The repository includes a `.github/workflows/ci.yml` file that will:
- ✅ Run tests on every push/PR
- ✅ Run linters
- ✅ Build the project
- ✅ Check code coverage

**To enable:**
1. Push the code to GitHub (the workflow file is already included)
2. GitHub Actions will automatically run on pushes and PRs
3. View results in the **"Actions"** tab

## 🤖 AI Assistant Capabilities

### What I Can Do:
- ✅ Read and analyze your codebase
- ✅ Create and modify files
- ✅ Run terminal commands (including git)
- ✅ Help you create issues/tasks via GitHub CLI
- ✅ Generate code, tests, and documentation
- ✅ Review code and suggest improvements

### What I Cannot Do Directly:
- ❌ I don't have direct GitHub API access (you need to authenticate)
- ❌ I can't push to your repository without your approval
- ❌ I can't create issues directly via GitHub API (but I can help you use `gh` CLI)

### How We Can Work Together:
1. **I create the content** (code, issues, tasks, docs)
2. **You review and approve** (via Cursor's interface)
3. **You commit and push** (or I can propose commands for you to run)
4. **I help you use GitHub CLI** to create issues/tasks

## 📝 Quick Reference Commands

```bash
# Check git status
git status

# Create a new branch
git checkout -b feature/your-feature-name

# Commit changes
git add .
git commit -m "Description of changes"

# Push to GitHub
git push origin your-branch-name

# Create a pull request
gh pr create --title "Your PR Title" --body "Description"

# List issues
gh issue list

# View an issue
gh issue view <number>

# Close an issue
gh issue close <number>

# Create a project board
gh project create --title "IAM Development" --format board
```

## 🔒 Security Best Practices

1. **Never commit secrets**: Use `.env` files (already in `.gitignore`)
2. **Use SSH keys** for GitHub authentication (more secure than HTTPS)
3. **Rotate tokens** regularly
4. **Use branch protection** rules in GitHub settings
5. **Review PRs** before merging

## 🆘 Troubleshooting

### "Remote origin already exists"
```bash
git remote remove origin
git remote add origin <your-repo-url>
```

### "Permission denied"
- Check your SSH keys: `ssh -T git@github.com`
- Or use HTTPS with personal access token

### "GitHub CLI not found"
- Install it: `sudo dnf install gh` (Fedora) or follow [GitHub CLI installation](https://cli.github.com/)

## 📚 Additional Resources

- [GitHub Docs](https://docs.github.com/)
- [GitHub CLI Docs](https://cli.github.com/manual/)
- [Git Best Practices](https://www.atlassian.com/git/tutorials/comparing-workflows)

---

**Next Steps:**
1. ✅ Create your GitHub repository
2. ✅ Connect it to this local repo
3. ✅ Push your code
4. ✅ Start creating issues and tasks!

Need help? Just ask! 🚀

