#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# This script deletes local branches that have already been merged to upstream.
#
# Usage: ./hack/scm/cleanup-branches.sh [--dry-run]
#
# Options:
#   --dry-run    Show which branches would be deleted without actually deleting them
#
# The script performs the following steps:
#   1. Fetches the latest changes from origin
#   2. Identifies branches that have been merged to the main branch
#   3. Deletes merged branches (excluding main, master, and current branch)

set -euo pipefail

# Color codes for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Protected branches that should never be deleted
readonly PROTECTED_BRANCHES=("main" "master" "develop" "dev")

DRY_RUN=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [--dry-run]"
      echo ""
      echo "Options:"
      echo "  --dry-run    Show which branches would be deleted without actually deleting them"
      echo "  -h, --help   Show this help message"
      exit 0
      ;;
    *)
      echo -e "${RED}Unknown option: $1${NC}"
      exit 1
      ;;
    esac
done

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         SPIKE SDK Go - Merged Branch Cleanup                   ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

if [ "$DRY_RUN" = true ]; then
  echo -e "${YELLOW}Running in DRY-RUN mode - no branches will be deleted${NC}"
  echo ""
fi

# Get current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo -e "${BLUE}Current branch: ${CURRENT_BRANCH}${NC}"
echo ""

# Step 1: Fetch latest from origin
echo -e "${YELLOW}Fetching latest changes from origin...${NC}"
git fetch origin --prune
echo -e "${GREEN}Done fetching.${NC}"
echo ""

# Step 2: Get merged branches
echo -e "${YELLOW}Finding branches merged to origin/main...${NC}"
echo ""

# Get list of merged branches (excluding HEAD and remote tracking refs)
MERGED_BRANCHES=$(git branch --merged origin/main --format='%(refname:short)' 2>/dev/null || true)

if [ -z "$MERGED_BRANCHES" ]; then
  echo -e "${GREEN}No merged branches found.${NC}"
  exit 0
fi

# Function to check if a branch is protected
is_protected() {
  local branch=$1
  for protected in "${PROTECTED_BRANCHES[@]}"; do
    if [ "$branch" = "$protected" ]; then
      return 0
    fi
  done
  return 1
}

# Step 3: Delete merged branches
DELETED_COUNT=0
SKIPPED_COUNT=0

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

for branch in $MERGED_BRANCHES; do
  # Skip if it's the current branch
  if [ "$branch" = "$CURRENT_BRANCH" ]; then
    echo -e "${YELLOW}Skipping current branch: ${branch}${NC}"
    SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
    continue
  fi

  # Skip if it's a protected branch
  if is_protected "$branch"; then
    echo -e "${YELLOW}Skipping protected branch: ${branch}${NC}"
    SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
    continue
  fi

  # Delete or show the branch
  if [ "$DRY_RUN" = true ]; then
    echo -e "${BLUE}Would delete: ${branch}${NC}"
  else
    if git branch -d "$branch" 2>/dev/null; then
      echo -e "${GREEN}Deleted: ${branch}${NC}"
    else
      # Try force delete if regular delete fails
      echo -e "${YELLOW}Branch ${branch} requires force delete, skipping (use git branch -D manually if needed)${NC}"
      SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
      continue
    fi
  fi
  DELETED_COUNT=$((DELETED_COUNT + 1))
done

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Summary
echo -e "${BLUE}Summary:${NC}"
if [ "$DRY_RUN" = true ]; then
  echo -e "  Branches to delete: ${GREEN}${DELETED_COUNT}${NC}"
else
  echo -e "  Branches deleted: ${GREEN}${DELETED_COUNT}${NC}"
fi
echo -e "  Branches skipped: ${YELLOW}${SKIPPED_COUNT}${NC}"
echo ""

if [ "$DRY_RUN" = true ] && [ "$DELETED_COUNT" -gt 0 ]; then
  echo -e "${YELLOW}Run without --dry-run to delete these branches.${NC}"
fi

echo -e "${GREEN}Done!${NC}"
