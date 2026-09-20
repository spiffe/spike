# Spec: Harden the lint gate and clear every finding

## Problem Statement

`make audit` ran golangci-lint with the default linter set, default
exclusion presets, errcheck's built-in ignore list, and no security
linter. Measured on 2026-09-20 with those suppressions removed, the
tree carries 396 findings: 344 errcheck, 34 gosec, 18 staticcheck.
Among them are real defects in production code: five silent
`strconv` failures in the sqlite secret loader, fourteen ignored
`pflag` getter errors, unchecked `Close`/`Remove` calls on files
holding secret material, and nine unchecked integer conversions on
shard indices and counts. SPIKE handles secrets; an error that is
neither handled, returned, nor logged is a bug, and a suppression
nobody assessed is a broken window.

## Proposed Solution

1. `.golangci.yml`: no exclusion presets, no path exclusions;
   `errcheck` with `disable-default-exclusions`, `check-blank`,
   `check-type-assertions`; `staticcheck` checks `all` (QF and ST
   included); `gosec` enabled. This is the gate from now on.
2. Fix every finding at the site. Rules:
   - Never suppress errcheck. Return the error, fail the test, or
     fail the process through the SDK `log.FatalLn`/`log.FatalErr`
     (never `os.Exit`, never bare `panic`).
   - A CLI write to stdout that fails is a failure: the caller must
     not exit 0 having printed nothing.
   - Tests use `t.Setenv` where a `*testing.T` exists and the test
     is not parallel; `TestMain` and helpers without `t` check the
     error and fail through `log.FatalLn`. Cleanup closes are checked
     via `t.Cleanup`.
   - gosec G115: add an explicit bounds check before the conversion
     and return an error when the value does not fit. G306: test
     fixtures are written with 0600. G304: clean the operator-supplied
     path with `filepath.Clean`. G101: the rule matches the identifier
     name against a credential pattern and then an entropy heuristic on
     the literal; when it misfires on a SQL statement constant, rename
     the constant after what the statement does. No `//nolint` at all:
     the first directive in a tree is the precedent the next one cites.
     No blanket exclusions in the config.
   - QF1008/QF1011/QF1012 and ST1000: apply the quick fix or add the
     package comment.
3. Files stay within the 80-column rule; comments use proper English.

## File Surface

- `.golangci.yml` (rewritten)
- Roughly 45 Go files across `app/nexus`, `app/spike`, `app/demo`,
  `app/bootstrap`, `app/keeper`, `internal/out`. The
  majority are test files (os.Setenv/Unsetenv and Close cleanup).

## Follow-through (2026-09-20)

The user rejected "pre-existing" as a reason to leave anything: the
project owns its quality. In the same change:

- Every Go line is within 80 columns (tab = 2), 203 lines rewrapped.
- Tracked Markdown prose is within 80 columns; only tables, fenced code,
  bare links, link definitions, and headings remain long. Context files
  were reflowed with `ctx fmt`.
- All `FIX-ME` comments are gone, each resolved in code: the shard
  response parser now rejects `null`, unknown fields, trailing data, and a
  keeper-reported error code (a `null` body used to hand a nil shard to
  the caller); an unreachable "invalid URL" case now uses a URL that
  really fails to parse; the remaining lite-backend and keeper-root cases
  are covered by `specs/pilot-failure-semantics.md`'s sibling work in the
  nexus packages.
- Doc comments follow the CONVENTIONS.md godoc shape; package comments
  live in `doc.go` files (`app/doc.go`, `internal/config/doc.go`,
  `app/spike/internal/cmd/format/doc.go`).
- Pilot exit codes and the entry point's exit are specified in
  `specs/pilot-failure-semantics.md`.
- A Go file never mixes exported and unexported functions or methods
  (public API in a small file, implementation in a `*_impl.go` sibling,
  test helpers in `test_helper_test.go`). The rule is in CONVENTIONS.md and
  enforced by `internal/layout/public_private_test.go`; the one
  production offender (`flags.go`) and the 17 test files carrying
  unexported helpers were split accordingly.
- The audit's golangci-lint step now runs with `--build-tags integration`,
  so the opt-in CLI harness is linted too; it carried five unchecked
  errors, now handled through `killProcess` and `stopProcess` helpers.
  gosec G204 (subprocess launched with variable) is excluded for that one
  package only, with the reason in `.golangci.yml`: a harness whose
  purpose is to exec the binary under test with per-test arguments can
  never satisfy G204 by construction. That is the single exclusion rule
  in the config, and it is assessed, scoped, and written down; there are
  still no `//nolint` directives.
- The route guard scans (Nexus and Keeper) no longer carry exempt lists.
  They recognize a handler by its signature, walk every package under
  `route/`, and follow same-package delegates transitively; the only
  skip is the router package itself, whose dispatcher shares the
  signature. Negative tests prove an unguarded handler is flagged, a
  delegating handler is accepted, and the walk sees handlers at all.
- Test helper files are `_test.go` files (`test_helper_test.go`), so no
  test scaffolding compiles into a binary.

## Non-Goals

- No behavior changes beyond error propagation and bounds checks.
- No `//nolint` directives, reasoned or not. A rename is acceptable only
  when the new name is at least as accurate as the old one.
- No changes to `go.mod` (owned by `specs/go-1-27-bump.md` and
  `specs/vuln-remediation.md` Round 3).

## Verification (2026-09-20, go1.27.1 linux/arm64, GOTOOLCHAIN=local)

- `make audit` exits 0 with the hardened config: golangci-lint
  reports 0 issues; govulncheck reports no called vulnerabilities.
- `make test` (race, uncached), final tree: 24 packages ok, 365 passes,
  0 failures. Tests were added for corrupt sqlite metadata rows, the
  flags helper, keeper ID parsing, keeper URL validation, the shard
  response parser, Pilot exit codes through cobra, and the diagnostics
  log routing.
- `grep -rn 'nolint' --include='*.go'` prints nothing.
- An independent review of the production diff found no blockers.
  Two of its notes were applied (bootstrap rejects keeper ID zero;
  the recovery out-of-range fatal zeroes its buffer first). One is
  left to the owner: the pre-existing `os.Exit(1)` in the Pilot's
  `Execute`, whose replacement changes the CLI's error output format.
