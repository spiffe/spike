#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Script to delete local branches where PRs have been closed or merged
# Requires GitHub CLI to be on the system and user logged in to GitHub
# via `gh auth login`.

set -e

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
  echo "Error: gh CLI is not installed."
  echo "Please install it from https://cli.github.com/"
  exit 1
fi

MAIN_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD | \
  sed 's@^refs/remotes/origin/@@')
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

echo "Main branch: $MAIN_BRANCH"
echo "Current branch: $CURRENT_BRANCH"
echo ""
echo "Fetching closed and merged PRs..."
echo ""

BRANCHES_TO_DELETE=()

# Get all closed and merged PRs and their head branch names
PR_BRANCHES=$(gh pr list --state merged --json headRefName \
  --jq '.[].headRefName' 2>/dev/null)
PR_BRANCHES+=$'\n'
PR_BRANCHES+=$(gh pr list --state closed --json headRefName \
  --jq '.[].headRefName' 2>/dev/null)

# Check each PR branch to see if it exists locally
while IFS= read -r branch; do
  if [ -z "$branch" ]; then
    continue
  fi

  # Skip main/master branches
  if [ "$branch" = "main" ] || [ "$branch" = "master" ] || \
     [ "$branch" = "$MAIN_BRANCH" ] || [ "$branch" = "$CURRENT_BRANCH" ]; then
    continue
  fi

  # Check if branch exists locally
  if git show-ref --verify --quiet "refs/heads/$branch"; then
    echo "Found local branch '$branch' with closed/merged PR"
    BRANCHES_TO_DELETE+=("$branch")
  fi
done <<< "$PR_BRANCHES"

if [ ${#BRANCHES_TO_DELETE[@]} -eq 0 ]; then
  echo "No local branches with closed or merged PRs found."
  exit 0
fi

echo ""
echo "The following branches have closed or merged PRs:"
for branch in "${BRANCHES_TO_DELETE[@]}"; do
  echo "  - $branch"
done
echo ""

read -p "Delete these branches? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

for branch in "${BRANCHES_TO_DELETE[@]}"; do
  echo "Deleting branch: $branch"
  git branch -D "$branch"
done

echo ""
echo "Done."
