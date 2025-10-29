#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# This script checks out a pull request locally for testing.

# Check if gh is available
if ! command -v gh &> /dev/null; then
  echo "Error: gh (GitHub CLI) is not installed or not in PATH"
  exit 1
fi

# Check if we are in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
  echo "Error: Not in a git repository"
  exit 1
fi

# Ask for PR number
echo "Enter the pull request number to checkout:"
read -r PR_NUMBER

# Validate PR number is a number
if ! [[ "$PR_NUMBER" =~ ^[0-9]+$ ]]; then
  echo "Error: Invalid PR number. Please enter a valid number."
  exit 1
fi

# Checkout the PR
echo "Checking out PR #$PR_NUMBER..."
gh pr checkout "$PR_NUMBER"

if [ $? -eq 0 ]; then
  echo "Successfully checked out PR #$PR_NUMBER"
  echo ""
  echo "To return to main branch, run: git checkout main"
else
  echo "Failed to checkout PR #$PR_NUMBER"
  exit 1
fi
