# Spec: Remediate govulncheck-reported vulnerabilities

## Problem Statement

`make audit` fails because `govulncheck` reports 10 vulnerabilities the
code actually calls, across the Go standard library and three modules.
The build used the locally installed Go 1.26.2 toolchain, which carries
unpatched stdlib advisories. This blocks the audit gate required before
any commit.

Called vulnerabilities:

- Standard library (Go 1.26.2): `net/textproto`, `mime`, `crypto/x509`
  (fixed in 1.26.4); `html/template` (x2), `net`, `net/http`
  (fixed in 1.26.3). IDs: GO-2026-5039, -5038, -5037, -4982, -4980,
  -4971, -4918 (net/http path).
- `golang.org/x/net` v0.48.0 (GO-2026-4918, GO-2026-4926).
- `github.com/go-jose/go-jose/v4` v4.1.3 (GO-2026-4945).
- `github.com/cloudflare/circl` v1.6.2 (GO-2026-4550).

## Proposed Solution

1. Pin the Go toolchain to 1.26.4 via a `toolchain go1.26.4` directive
   in `go.mod` (keeps the `go 1.25.5` language baseline; builds use the
   patched stdlib). Resolves all standard-library advisories.
2. Upgrade the three flagged modules to their fixed versions:
   - `github.com/cloudflare/circl` v1.6.2 -> v1.6.3
   - `github.com/go-jose/go-jose/v4` v4.1.3 -> v4.1.4
   - `golang.org/x/net` v0.48.0 -> v0.55.0
3. `go mod tidy` to settle the module graph (transitively bumps
   `golang.org/x/crypto`, `x/sys`, `x/term`, `x/text`).

## File Surface

- `go.mod` (modified): `toolchain go1.26.4`; bumped requires.
- `go.sum` (modified): checksums for the upgraded graph.

## Error / Edge Cases

- **Toolchain availability.** `go1.26.4` is fetched via Go's toolchain
  mechanism (`GOTOOLCHAIN`); contributors on older Go auto-download it
  because of the `toolchain` directive.
- **Remaining uncalled vulns.** govulncheck still lists vulnerabilities
  in imported/required packages that the code does not call. These do
  not fail the gate and are out of scope here; they resolve naturally as
  those deps are bumped over time.

## Non-Goals

- No application-code changes; remediation is dependency/toolchain only.
- Not upgrading the `go` language directive beyond 1.25.5.

## Verification

- `make audit` exits 0; `govulncheck` reports "Your code is affected by
  0 vulnerabilities."
- `make test` passes on the upgraded module graph (Go 1.26.4).
