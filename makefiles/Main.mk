#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
# \\\\\\

confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

no-dirty:
	@test -z "$(shell git status --porcelain)"

upgradeable:
	@go run github.com/oligot/go-mod-upgrade@latest

tidy:
	go mod tidy -v
	go fmt ./...