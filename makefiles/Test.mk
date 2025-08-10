#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
# \\\\\\

.PHONY: lint-go
lint-go:
	./hack/qa/lint-go.sh

test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

test:
	go test -v -race -buildvcs ./...

audit: test
	go mod tidy -diff
	go mod verify
	test -z "$(shell gofmt -l .)"
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...