#!/bin/bash

# Script to update GitHub Project Kanban board
# This script updates project items status and links commits

PROJECT_NUMBER=1
USER="cvs0986"

echo "Updating GitHub Project Kanban Board..."

# Get project ID
PROJECT_ID=$(gh api graphql -f query="
{
  user(login: \"$USER\") {
    projectV2(number: $PROJECT_NUMBER) {
      id
    }
  }
}" | jq -r '.data.user.projectV2.id')

echo "Project ID: $PROJECT_ID"

# Get Status field ID and options
STATUS_FIELD=$(gh api graphql -f query="
{
  user(login: \"$USER\") {
    projectV2(number: $PROJECT_NUMBER) {
      fields(first: 20) {
        nodes {
          ... on ProjectV2SingleSelectField {
            id
            name
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}" | jq -r '.data.user.projectV2.fields.nodes[] | select(.name == "Status") | .id')

echo "Status Field ID: $STATUS_FIELD"

# Get "Done" option ID
DONE_OPTION=$(gh api graphql -f query="
{
  user(login: \"$USER\") {
    projectV2(number: $PROJECT_NUMBER) {
      fields(first: 20) {
        nodes {
          ... on ProjectV2SingleSelectField {
            id
            name
            options {
              id
              name
            }
          }
        }
      }
    }
  }
}" | jq -r '.data.user.projectV2.fields.nodes[] | select(.name == "Status") | .options[] | select(.name == "Done" or .name == "In Progress" or .name == "Complete") | "\(.id)|\(.name)"')

echo "Done Option: $DONE_OPTION"

# Map issues to project items and update status
# Issues 1-9 are completed
for ISSUE_NUM in 1 2 3 4 5 6 7 8 9; do
    echo "Processing Issue #$ISSUE_NUM..."
    
    # Get issue ID
    ISSUE_ID=$(gh api graphql -f query="
    {
      repository(owner: \"$USER\", name: \"ARauth\") {
        issue(number: $ISSUE_NUM) {
          id
        }
      }
    }" | jq -r ".data.repository.issue.id")
    
    if [ "$ISSUE_ID" != "null" ] && [ -n "$ISSUE_ID" ]; then
        echo "  Issue ID: $ISSUE_ID"
        
        # Get project item ID for this issue
        ITEM_ID=$(gh api graphql -f query="
        {
          user(login: \"$USER\") {
            projectV2(number: $PROJECT_NUMBER) {
              items(first: 20) {
                nodes {
                  id
                  content {
                    ... on Issue {
                      id
                      number
                    }
                  }
                }
              }
            }
          }
        }" | jq -r ".data.user.projectV2.items.nodes[] | select(.content.id == \"$ISSUE_ID\") | .id")
        
        if [ -n "$ITEM_ID" ] && [ "$ITEM_ID" != "null" ]; then
            echo "  Item ID: $ITEM_ID"
            
            # Update status to "Done" (use "Done" option specifically)
            DONE_ID=$(echo "$DONE_OPTION" | grep "|Done$" | cut -d'|' -f1)
            DONE_NAME="Done"
            
            # If "Done" not found, use the last option (usually Done)
            if [ -z "$DONE_ID" ] || [ "$DONE_ID" == "null" ]; then
                DONE_ID=$(echo "$DONE_OPTION" | tail -1 | cut -d'|' -f1)
                DONE_NAME=$(echo "$DONE_OPTION" | tail -1 | cut -d'|' -f2)
            fi
            
            if [ -n "$DONE_ID" ] && [ "$DONE_ID" != "null" ]; then
                echo "  Updating status to: $DONE_NAME"
                gh api graphql -f query="
                mutation {
                  updateProjectV2ItemFieldValue(
                    input: {
                      projectId: \"$PROJECT_ID\"
                      itemId: \"$ITEM_ID\"
                      fieldId: \"$STATUS_FIELD\"
                      value: {
                        singleSelectOptionId: \"$DONE_ID\"
                      }
                    }
                  ) {
                    projectV2Item {
                      id
                    }
                  }
                }" > /dev/null 2>&1
                
                if [ $? -eq 0 ]; then
                    echo "  ✅ Updated Issue #$ISSUE_NUM status"
                else
                    echo "  ⚠️  Failed to update Issue #$ISSUE_NUM"
                fi
            fi
        fi
    fi
done

echo ""
echo "Project Kanban update complete!"

