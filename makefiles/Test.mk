#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Run Go linting using custom script
# Usage: make lint-go
# Executes the project's Go linting script located in hack/qa/
# Depends on ./hack/qa/lint-go.sh being present and executable
.PHONY: lint-go
lint-go:
	./hack/qa/lint-go.sh

# Run tests with coverage report and open HTML visualization
# Usage: make test/cover
# Executes all tests with race detection and coverage profiling
# Generates an HTML coverage report and opens it in the default browser
# Coverage data is temporarily stored in /tmp/coverage.out
# Flags: -v (verbose), -race (race detection), -buildvcs (include VCS info)
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

# Run all tests with race detection
# Usage: make test
# Executes all tests in the project with verbose output and race detection
# Does not generate coverage reports (use test/cover for that)
# Flags: -v (verbose), -race (race detection), -buildvcs (include VCS info)
# Packages run concurrently (Go's default -p). This is safe because the
# sqlite-backed test packages isolate their data directories per run via
# TestMain (SPIKE_NEXUS_DATA_DIR points at a temporary directory), no
# test binds a fixed network port, and environment variables are
# per-process, so package-level parallelism cannot leak state.
.PHONY: test
test:
	go test -v -race -buildvcs ./...

# Comprehensive code quality audit
# Usage: make audit
# Prerequisite: runs 'test' target first to ensure tests pass
# Performs multiple quality checks:
#   1. go mod tidy -diff: checks if go.mod needs tidying
#      (fails if changes needed)
#   2. go mod verify: verifies module dependencies haven't been tampered with
#   3. gofmt check: ensures all Go files are properly formatted
#   4. go vet: runs Go's built-in static analysis
#   5. staticcheck: runs advanced static analysis
#      (excluding ST1000, U1000 checks)
#   6. govulncheck: scans for known security vulnerabilities
#   7. golangci-lint: runs a comprehensive set of linters
#      (follows the configuration in .golangci.yml)
.PHONY: audit
audit:
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	CGO_ENABLED=0 go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

# Comprehensive set of checks to simulate a CI environment
# Usage: make ci
# Prerequisites:
#   1. runs 'test' target first to ensure tests pass
#   2. runs 'audit' target to perform code quality checks
.PHONY: ci
ci: test audit

# Run the live, opt-in Pilot integration suite (specs/integration-tests.md,
# Slice B). Requires a healthy `make start` environment (SPIRE, Nexus, and
# the Keepers) and the spike binary on PATH. Non-destructive: it exercises a
# CRUD and cipher smoke pass plus the Nexus-unreachable warning. The
# integration build tag and SPIKE_INTEGRATION_TEST=1 keep these tests out of
# the normal `make test`.
# Usage: make integration-test
.PHONY: integration-test
integration-test:
	SPIKE_INTEGRATION_TEST=1 go test -tags=integration -count=1 -v \
		./app/spike/internal/cmd/integration/...

# Run the integration suite including the DESTRUCTIVE uninitialized-Nexus
# case, which kills Nexus and every Keeper and restarts Nexus alone. It
# leaves the environment uninitialized (but cleans up the Nexus it spawned):
# reset it with Ctrl+C on the make start terminal (or make kill), then
# make start.
# Usage: make integration-test-destructive
.PHONY: integration-test-destructive
integration-test-destructive:
	SPIKE_INTEGRATION_TEST=1 SPIKE_INTEGRATION_DESTRUCTIVE=1 \
		go test -tags=integration -count=1 -v \
		./app/spike/internal/cmd/integration/...
