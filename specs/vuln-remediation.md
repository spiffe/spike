# Spec: Remediate govulncheck-reported vulnerabilities

Advisories land on their own schedule, so this is a recurring exercise
rather than a one-off. Each round is recorded below with the advisory
that triggered it and what was done.

# Round 1: toolchain and dependency sweep (2026-06-13)

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
4. Clear the remaining (uncalled) advisories surfaced after the above:
   bump `golang.org/x/crypto` v0.51.0 -> v0.52.0 (13 advisories) and
   `google.golang.org/grpc` v1.78.0 -> v1.79.3 (GO-2026-4762); tidy.
   govulncheck then reports zero vulnerabilities total, not merely zero
   called.

## File Surface

- `go.mod` (modified): `toolchain go1.26.4`; bumped requires.
- `go.sum` (modified): checksums for the upgraded graph.

## Error / Edge Cases

- **Toolchain availability.** `go1.26.4` is fetched via Go's toolchain
  mechanism (`GOTOOLCHAIN`); contributors on older Go auto-download it
  because of the `toolchain` directive.
- **Uncalled advisories are still remediated.** "Uncalled" means lower
  urgency, not acceptable. The x/crypto and grpc bumps in step 4 clear
  them so the scan is clean end to end, not just on called paths.

## Non-Goals

- No application-code changes; remediation is dependency/toolchain only.
- Not upgrading the `go` language directive beyond 1.25.5.

## Verification

- `make audit` exits 0; `govulncheck` reports "No vulnerabilities
  found" with no "also found in packages you import / modules you
  require" residue (0 vulnerabilities total).
- `make test` passes on the upgraded module graph (Go 1.26.4), including
  the grpc v1.79.3 minor bump.

---

# Round 2: GO-2026-5970 (2026-07-25)

## Problem Statement

CI on `main` began failing at the "Go - Lint" job, which runs
`make audit`. The failure is not caused by any code change: run #300
(2026-07-18) passed and run #301 (2026-07-25) failed, but the merge in
between never touched `go.mod` or `go.sum`, and `golang.org/x/text`
was `v0.37.0` on both sides. A new advisory was published in the
interval, so the build broke on a timer rather than on a commit.

`GO-2026-5970` (infinite loop on invalid input in `golang.org/x/text`)
is **called** code, reached from
`recovery.sendShardsToKeepers` -> `net.Post` -> `norm.Form.*`, so
govulncheck exits non-zero.

## Proposed Solution

1. `golang.org/x/text` v0.37.0 -> v0.39.0 (fixes GO-2026-5970).
2. `golang.org/x/net` v0.55.0 -> v0.56.0 (clears the uncalled
   GO-2026-5942 in the same pass).
3. `go mod tidy` to settle the graph, which transitively lifts
   `x/crypto` v0.52.0 -> v0.53.0, `x/sys` v0.45.0 -> v0.46.0, and
   `x/term` v0.43.0 -> v0.44.0.

## File Surface

- `go.mod`, `go.sum` (modified). No application-code changes.

## Error / Edge Cases

- **The "0 vulnerabilities total" bar from Round 1 is no longer
  reachable.** `GO-2026-5932` reports that
  `golang.org/x/crypto/openpgp` is unmaintained and unsafe by design,
  with `Fixed in: N/A`. No version bump can clear it: the package is
  deprecated rather than patched. It arrives transitively and SPIKE
  does not call it, so govulncheck exits 0 and CI passes with it
  present.

  Round 1's verification criterion is therefore amended, not merely
  missed: the standing bar is **zero called vulnerabilities**, plus a
  recorded justification for every uncalled one left behind. Clearing
  uncalled findings remains preferred where a fixed version exists.

## Non-Goals

- Not vendoring or forking `x/crypto` to excise `openpgp`. It is
  uncalled; the cost outweighs the benefit until an upstream drops it.
- No application-code changes.

## Verification

- `make audit` exits 0, including `go mod tidy -diff`, `go vet`,
  staticcheck, govulncheck, and the `CGO_ENABLED=0` golangci-lint run.
- `govulncheck ./...` reports "No vulnerabilities found" under Symbol
  Results; the only residue is the unfixable GO-2026-5932 under Module
  Results.
- `make test` passes on the upgraded graph; `go build ./...` clean.
