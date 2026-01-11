#!/bin/bash

# Script to create GitHub issues for Capability Model Implementation
# Requires GitHub CLI (gh) to be installed and authenticated

set -e

REPO="nuage-indentity"  # Update with your actual repo name
BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ISSUES_FILE="${BASE_DIR}/docs/planning/GITHUB_ISSUES.md"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Creating GitHub issues for Capability Model Implementation...${NC}"

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

# Function to create an issue from markdown
create_issue() {
    local issue_num=$1
    local title=$2
    local description=$3
    local labels=$4
    
    echo -e "${YELLOW}Creating issue #${issue_num}: ${title}${NC}"
    
    # Create issue using gh CLI
    gh issue create \
        --title "${title}" \
        --body "${description}" \
        --label "${labels}" \
        --repo "${REPO}" || {
        echo -e "${RED}Failed to create issue #${issue_num}${NC}"
        return 1
    }
    
    echo -e "${GREEN}✓ Issue #${issue_num} created${NC}"
}

# Parse issues from GITHUB_ISSUES.md and create them
# This is a simplified version - you may need to adjust based on your markdown structure

echo -e "${GREEN}Reading issues from ${ISSUES_FILE}...${NC}"

# Note: This script provides a template. You'll need to manually create issues
# or enhance this script to parse the markdown file properly.

echo -e "${YELLOW}Note: This script is a template.${NC}"
echo -e "${YELLOW}To create issues, either:${NC}"
echo -e "  1. Manually create issues using the GitHub web UI"
echo -e "  2. Use the gh CLI interactively"
echo -e "  3. Enhance this script to parse GITHUB_ISSUES.md"

# Example: Create a single issue to test
# create_issue "001" \
#     "[Phase 1] Create tenant_capabilities table" \
#     "Create database migration and table for storing which capabilities are allowed for each tenant." \
#     "database,migration,p0,capability-model,phase-1"

echo -e "${GREEN}Done!${NC}"
echo -e "${YELLOW}See ${ISSUES_FILE} for all issue details.${NC}"




