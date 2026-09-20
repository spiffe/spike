# Spec: Bump the Go requirement to 1.27.1

## Problem Statement

`go.mod` requires Go 1.26.5. Go 1.27.1 (linux/arm64) is installed and
verified on the development host, and the module should track it so
CI (`go-version-file: go.mod`), the Docker builder images, and local
builds all use one toolchain. Two consequences must land in the same
change or a CI gate turns red:

- The builder images pin `golang:1.26.5` and ship with
  `GOTOOLCHAIN=local`, so they refuse a newer `go` directive (the
  same failure #303 fixed for 1.26.5).
- golangci-lint v1.64.8, the final v1 release and the version `make
  audit` resolved via `@latest`, cannot read Go 1.27 export data
  (`export data version 4 is greater than maximum supported version
  2`). Every package then fails typecheck.

Validation also surfaced two called govulncheck advisories in
`google.golang.org/grpc` v1.79.3 that are present on the base commit.
Their remediation was authorized mid-change and is specified in
`specs/vuln-remediation.md`, Round 3; this spec only lists the
resulting module changes.

## Proposed Solution

1. `go.mod`: `go 1.26.5` -> `go 1.27.1`. No `toolchain` directive
   (none existed; the local toolchain equals the requirement).
2. `dockerfiles/*.Dockerfile` (5 files): `golang:1.26.5` ->
   `golang:1.27.1`. The tag is published on Docker Hub.
3. `makefiles/Test.mk`: audit step 7 runs
   `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`.
4. `.golangci.yml`: migrated to the v2 format with lint scope parity
   (see below).
5. Dependency remediation per Round 3: `google.golang.org/grpc`
   v1.79.3 -> v1.83.2 and `golang.org/x/crypto` v0.53.0 -> v0.56.0,
   then `go mod tidy`. Transitive consequences: `go-spiffe/v2`
   v2.6.0 -> v2.7.0 (grpc's own requirement), `x/net` v0.58.0,
   `x/sys` v0.47.0, `x/term` v0.45.0, `x/text` v0.41.0, `x/oauth2`
   v0.36.0, and a newer `genproto/googleapis/rpc` pseudo-version.

## Lint Scope Parity

Original (v1) selection:

```yaml
linters:
  enable: [errcheck, gosimple, govet, ineffassign, staticcheck, unused]
```

plus golangci-lint v1 default exclusions.

New (v2) selection:

- Linters: the v2 default set, which is exactly `errcheck`, `govet`,
  `ineffassign`, `staticcheck`, `unused`. `gosimple` no longer exists
  as a separate linter; its checks are part of `staticcheck`.
- Exclusions: presets `comments`, `common-false-positives`, `legacy`,
  `std-error-handling` and the `third_party$`/`builtin$`/`examples$`
  paths, as emitted by `golangci-lint migrate`. They reproduce the v1
  default exclusion rules.
- staticcheck checks: the v2 default list (`all` minus ST1000, ST1003,
  ST1016, ST1020, ST1021, ST1022) plus `-QF*`. The QF quick-fix
  analyzers were never run by the v1 `staticcheck`/`gosimple` linters.
  Enabling them reports 7 findings in code this change does not touch
  (QF1008 x3 in `route/secret/map_test.go`, QF1011 x1 in
  `backend/lite/initialize_test.go`, QF1012 x3 in
  `cmd/policy/format.go`), so they are excluded to keep the scope
  identical. Enabling QF checks is a separate, deliberate change.

## File Surface

- `go.mod`, `go.sum` (modified)
- `dockerfiles/bootstrap.Dockerfile`, `demo.Dockerfile`,
  `keeper.Dockerfile`, `nexus.Dockerfile`, `pilot.Dockerfile`
  (modified, one line each)
- `makefiles/Test.mk` (modified, one line)
- `.golangci.yml` (rewritten for v2)
- `specs/vuln-remediation.md` (Round 3), `.context/TASKS.md`,
  `.context/DECISIONS.md`, `.context/LEARNINGS.md` (records)

## Non-Goals

- No blanket dependency upgrade; only the modules named in Round 3
  and what `go mod tidy` lifts with them.
- No application-code changes, including fixes for the QF findings.
- No change to the standalone staticcheck step (`-checks=all,-ST1000,
  -U1000`), which passes on 1.27.1 as is.

## Verification (2026-09-19, go1.27.1 linux/arm64, GOTOOLCHAIN=local)

- `go mod tidy -diff`: empty. `go build ./...` and `make build`: pass.
- `make test` (race detector), on the final module graph: 22 packages
  ok, 350 passes, 0 failures.
- `make audit`, on the final module graph: exit 0. govulncheck
  reports 0 called and 0 package-level vulnerabilities; the only
  residue is the unfixable GO-2026-5932 module notice (see Round 2).
  golangci-lint v2: 0 issues.
- Before the Round 3 bumps, govulncheck output on this tree and on
  base commit 42547a7e (same toolchain) was byte-identical, which is
  how the findings were classified as baseline.
- Not run: Docker image builds, integration tests, live drills.
