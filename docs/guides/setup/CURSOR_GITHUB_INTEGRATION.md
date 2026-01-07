# 🔗 Connecting GitHub to Cursor

Yes! You can connect your GitHub account directly to Cursor for seamless development workflows. This guide will walk you through the process.

## ✅ What You Can Do After Connecting

Once connected, you'll be able to:
- ✅ **Access repositories** directly from Cursor
- ✅ **Create and manage issues** without leaving the IDE
- ✅ **View pull requests** and manage them
- ✅ **Sync code** seamlessly
- ✅ **Use GitHub Copilot** features (if available)
- ✅ **Access repository context** for better AI assistance

## 🚀 Step-by-Step Connection Guide

### Method 1: Through Cursor Settings (Recommended)

1. **Open Cursor Settings**:
   - Press `Ctrl + ,` (or `Cmd + ,` on Mac)
   - Or go to: **File → Preferences → Settings**

2. **Navigate to Integrations**:
   - In the settings search bar, type: `github`
   - Or look for **"Integrations"** in the left sidebar
   - Click on **"GitHub"** or **"Integrations"**

3. **Connect GitHub Account**:
   - Click the **"Connect"** button next to GitHub
   - You'll be redirected to GitHub's authorization page
   - Sign in to GitHub if prompted

4. **Authorize Cursor**:
   - Review the permissions Cursor is requesting
   - Choose repository access:
     - **All repositories** (recommended for full access)
     - **Selected repositories** (more restrictive)
   - Click **"Authorize"** or **"Install"**

5. **Verify Connection**:
   - Return to Cursor
   - The connection status should show as **"Connected"**
   - You may see your GitHub username/avatar

### Method 2: Through Cursor Dashboard

1. **Open Cursor Dashboard**:
   - Look for a dashboard icon or menu
   - Or go to: **View → Command Palette** → Type "Dashboard"

2. **Find Integrations Section**:
   - Navigate to the **"Integrations"** tab
   - Look for **GitHub** in the list

3. **Connect**:
   - Click **"Connect"** next to GitHub
   - Follow the same authorization steps as Method 1

### Method 3: Through Command Palette

1. **Open Command Palette**:
   - Press `Ctrl + Shift + P` (or `Cmd + Shift + P` on Mac)

2. **Search for GitHub**:
   - Type: `GitHub: Connect`
   - Select the command to connect your account

3. **Follow Authorization**:
   - Complete the GitHub authorization flow

## 🔍 Verifying Your Connection

After connecting, verify it worked:

1. **Check Settings**:
   - Go to Settings → Integrations
   - GitHub should show as "Connected"

2. **Test Repository Access**:
   - Try opening a GitHub repository
   - Or use: `Ctrl + Shift + P` → `GitHub: Clone Repository`

3. **Check Account Info**:
   - Look for your GitHub username/avatar in Cursor's status bar or settings

## 🛠️ Troubleshooting

### Issue: "Connect" Button Still Shows After Authorizing

**Solution**:
- Sometimes the UI doesn't update immediately
- Check your account settings in Cursor
- Try disconnecting and reconnecting
- Restart Cursor if needed

### Issue: Authorization Page Doesn't Open

**Solution**:
- Check your browser popup blocker
- Try opening the authorization URL manually
- Ensure you're signed into GitHub in your browser

### Issue: "Permission Denied" Errors

**Solution**:
- Go to GitHub → Settings → Applications → Authorized OAuth Apps
- Find "Cursor" and check its permissions
- Revoke and re-authorize if needed
- Ensure you granted access to the repositories you need

### Issue: Can't See Repositories

**Solution**:
- Verify you selected "All repositories" or the correct ones
- Check if the repositories are private (may need additional permissions)
- Try refreshing the connection

## 🔐 Security & Permissions

### What Permissions Does Cursor Request?

Cursor typically requests:
- ✅ **Read access** to your repositories
- ✅ **Write access** to create issues, PRs, etc.
- ✅ **Read access** to your profile information

### Best Practices:

1. **Review Permissions**: Only grant what you need
2. **Use Selected Repositories**: If you don't need access to all repos
3. **Regular Audits**: Periodically review authorized apps in GitHub settings
4. **Revoke When Done**: Disconnect if you no longer use Cursor

## 📋 After Connection: What's Next?

Once connected, you can:

### 1. Clone Repositories
```bash
# Via Command Palette
Ctrl + Shift + P → "GitHub: Clone Repository"
```

### 2. Create Issues
- Right-click in code → "Create GitHub Issue"
- Or use Command Palette: `GitHub: Create Issue`

### 3. View Pull Requests
- Open the Source Control panel
- Look for PR-related options
- Or use: `GitHub: View Pull Requests`

### 4. Sync Your Current Project

If you want to connect this project (`nuage-indentity`) to GitHub:

```bash
# If you haven't already
cd /home/eshwar/Documents/Veer/nuage-indentity

# Add remote (replace YOUR_USERNAME)
git remote add origin https://github.com/YOUR_USERNAME/nuage-identity.git

# Push to GitHub
git push -u origin main
```

## 🎯 Quick Commands Reference

After connecting, these commands become available:

| Command | Shortcut | Description |
|---------|----------|-------------|
| `GitHub: Clone Repository` | `Ctrl+Shift+P` | Clone a repo from GitHub |
| `GitHub: Create Issue` | `Ctrl+Shift+P` | Create a new issue |
| `GitHub: View Pull Requests` | `Ctrl+Shift+P` | View PRs for current repo |
| `GitHub: Open Repository` | `Ctrl+Shift+P` | Open repo in browser |
| `GitHub: Sign Out` | `Ctrl+Shift+P` | Disconnect GitHub account |

## 🔄 Disconnecting GitHub

If you need to disconnect:

1. Go to **Settings → Integrations → GitHub**
2. Click **"Disconnect"** or **"Sign Out"**
3. Confirm the disconnection
4. Optionally revoke access in GitHub Settings → Applications

## 📚 Additional Resources

- [Cursor Documentation](https://docs.cursor.com/)
- [GitHub Integration Docs](https://docs.cursor.com/integrations/github)
- [GitHub OAuth Apps](https://github.com/settings/applications)

## ❓ Still Having Issues?

If you're still having trouble connecting:

1. **Check Cursor Version**: Ensure you're on the latest version
2. **Check Internet Connection**: GitHub requires internet access
3. **Try Alternative Method**: Use GitHub CLI (`gh auth login`) as a backup
4. **Contact Support**: Reach out to Cursor support if issues persist

---

**Next Steps:**
1. ✅ Connect your GitHub account using one of the methods above
2. ✅ Verify the connection works
3. ✅ Connect this project to your GitHub repository
4. ✅ Start using GitHub features in Cursor!

Need help? Just ask! 🚀

