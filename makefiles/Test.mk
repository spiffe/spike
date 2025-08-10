#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
# \\\\\\

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
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

# Run all tests with race detection
# Usage: make test
# Executes all tests in the project with verbose output and race detection
# Does not generate coverage reports (use test/cover for that)
# Flags: -v (verbose), -race (race detection), -buildvcs (include VCS info)
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
# Designed for CI/CD pipelines to ensure code quality and security
audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
