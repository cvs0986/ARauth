# 📦 Install GitHub CLI - Quick Guide

## Step 1: Install GitHub CLI

Open your terminal and run:

```bash
sudo dnf install -y gh
```

Enter your sudo password when prompted.

## Step 2: Authenticate GitHub CLI

After installation, authenticate:

```bash
gh auth login
```

You'll be prompted to:
1. **What account do you want to log into?** → Choose `GitHub.com`
2. **What is your preferred protocol?** → Choose `HTTPS` (recommended) or `SSH`
3. **Authenticate Git credential helper?** → Choose `Yes`
4. **How would you like to authenticate?** → Choose one:
   - **Login with a web browser** (easiest - will open browser)
   - **Paste an authentication token** (if you have a token)

If you choose web browser, it will:
- Open your browser
- Ask you to authorize GitHub CLI
- Return to terminal when done

## Step 3: Verify Installation

```bash
gh --version
gh auth status
```

You should see:
- GitHub CLI version
- Your authenticated account (cvs0986)

## Step 4: Test Connection

```bash
cd /home/eshwar/Documents/Veer/nuage-indentity
gh repo view cvs0986/ARauth
```

This should show your repository information.

## ✅ After Installation

Once GitHub CLI is installed and authenticated, I can:
- ✅ Create issues directly on GitHub
- ✅ Create tasks with templates
- ✅ Set up GitHub Projects/Kanban boards
- ✅ Create and manage branches
- ✅ Create pull requests
- ✅ Manage your entire development workflow

---

**Run the commands above, then let me know when it's done!** 🚀

