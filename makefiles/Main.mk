#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Interactive confirmation prompt
# Usage: make confirm && make some-destructive-action
# Prompts user for confirmation before proceeding with potentially destructive
# operations
# Returns successfully only if user explicitly types 'y', defaults to 'N' on
# empty input
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# Check for uncommitted changes in git repository
# Usage: make no-dirty
# Ensures the working directory is clean with no uncommitted changes
# Useful as a prerequisite for deployment or release targets
# Exits with error code if there are any modified, added, or untracked files
no-dirty:
	@test -z "$(shell git status --porcelain)"

# Check for available Go module upgrades
# Usage: make upgradeable
# Downloads and runs go-mod-upgrade tool to display available dependency updates
# Does not actually upgrade anything, only shows what could be upgraded
# Requires internet connection to fetch the tool and check for updates
upgradeable:
	@go run github.com/oligot/go-mod-upgrade@latest

# Clean up Go module dependencies and format code
# Usage: make tidy
# Performs two operations:
#   1. go mod tidy -v: removes unused dependencies and adds missing ones
#   2. go fmt ./...: formats all Go source files in the project
# Should be run before committing code changes
tidy:
	go mod tidy -v
	go fmt ./...

# Create a signed git tag using version from app/VERSION.txt
# Usage: make tag
# Reads version from app/VERSION.txt and creates a signed tag with "v" prefix
# Example: if VERSION.txt contains "0.4.4", creates tag "v0.4.4"
tag:
	./hack/scm/tag.sh

.PHONY: docs
docs:
	./hack/bare-metal/build/build-docs.sh
	./hack/qa/cover.sh