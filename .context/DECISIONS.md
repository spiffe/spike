# Decisions

<!-- INDEX:START -->
| Date | Decision |
|----|--------|
| 2026-06-13 | Pin Go toolchain to 1.26.4 and bump circl/go-jose/x/net to clear govulncheck |
| 2026-06-13 | Own the ctx Makefile fragment as makefiles/Ctx.mk |
<!-- INDEX:END -->

<!-- DECISION FORMATS

## Quick Format (Y-Statement)

For lightweight decisions, a single statement suffices:

> "In the context of [situation], facing [constraint], we decided for [choice]
> and against [alternatives], to achieve [benefit], accepting that [trade-off]."

## Full Format

For significant decisions:

## [YYYY-MM-DD] Decision Title

**Status**: Accepted | Superseded | Deprecated

**Context**: What situation prompted this decision? What constraints exist?

**Alternatives Considered**:
- Option A: [Pros] / [Cons]
- Option B: [Pros] / [Cons]

**Decision**: What was decided?

**Rationale**: Why this choice over the alternatives?

**Consequence**: What are the implications? (Include both positive and negative)

**Related**: See also [other decision] | Supersedes [old decision]

## When to Record a Decision

✓ Trade-offs between alternatives
✓ Non-obvious design choices
✓ Choices that affect architecture
✓ "Why" that needs preservation

✗ Minor implementation details
✗ Routine maintenance
✗ Configuration changes
✗ No real alternatives existed

-->
## [2026-09-20-091449] Guard scan: handlers by signature, no exempt lists

**Status**: Accepted

**Context**: The AST test that proves every route handler invokes a guard
carried three by-name lists: seven 'utility' file names skipped before parsing,
four cipher helper names hard-coded as 'known guarded', and a fixed list of
route subdirectories; it also recognized handlers only by the Route name prefix
and skipped *_intercept.go by suffix. The maintainer asked whether a
grandfathered array was being kept, since any agent could append to it.

**Decision**: Route guard scan recognizes handlers by signature and follows
delegates; no exempt lists

**Rationale**: Rewrote both scans (Nexus and Keeper are now identical). A
handler is any function with the parameters (http.ResponseWriter, *http.Request,
*journal.AuditEntry); every package directory under route/ is walked, only the
router package itself is skipped because its Route dispatcher shares the
signature; a handler is guarded when it reaches ReadParseAndGuard or a
guard*/…Guard… function directly or through same-package functions it calls
or passes as values, followed transitively with a visited set. Negative tests
prove an unguarded handler is flagged and a delegating handler is accepted; a
third test fails if the walk sees zero handlers.

**Consequence**: There is no file, name, or directory list left to extend.
Renaming a handler, moving it to a new file, or adding a new route package
cannot bypass the scan. New guard helpers must follow the guard* naming contract
or call one that does.

---

## [2026-09-20-091047] Test helper files are _test.go files

**Status**: Accepted

**Context**: The house pattern kept per-package test scaffolding in a plain
test_helper.go, a non-_test file compiled into the production package; several
of them imported the testing package. Idiomatic Go keeps such helpers in
_test.go files, which only build under go test.

**Decision**: Test helper files are _test.go files

**Rationale**: The maintainer chose the idiomatic form. All 19 helper files were
renamed to test_helper_test.go (testing_helper_test.go in sqlite/persist),
keeping the base name for continuity and grep-ability. Build tags are preserved
(the integration harness helper keeps its tag). The guard scan's explicit
test_helper.go skip entry was dropped because _test.go files are already skipped
by suffix.

**Consequence**: No test scaffolding or testing import ships in a SPIKE binary.
The layout test still sees only unexported functions in these files.
CONVENTIONS.md names the new file pattern.

---

## [2026-09-20-090836] Audit lints the integration tag; G204 scoped to it

**Status**: Accepted

**Context**: The Pilot CLI integration harness
(app/spike/internal/cmd/integration, build tag integration) was never linted:
make audit did not set the tag. Under the tag it had 5 unchecked errors and 3
gosec G204 findings. G204 fires whenever a subprocess command or its arguments
are variables; an exec harness that runs the binary under test with per-test
arguments cannot satisfy that by construction.

**Decision**: Lint the integration-tagged harness in make audit; scope gosec
G204 to it

**Rationale**: Widened the audit step to --build-tags integration so the package
is gated like everything else, fixed the five errcheck sites (a pkill helper
that tolerates 'no match' but fails when pkill cannot run, and a stop helper
that reports a kill or reap failure), and added one exclusion rule to
.golangci.yml: gosec G204, path app/spike/internal/cmd/integration/ only, with
the reason written next to it. Chosen over three //nolint lines (a precedent)
and over leaving the package unlinted (a larger gap than G204).

**Consequence**: make audit covers the harness. The config's only exclusion is
that one rule in that one package; the tree still has zero nolint directives.
Adding any exec-based test elsewhere must be assessed on its own, not by
widening the rule.

---

## [2026-09-20-090127] Enforce the public/private file split by AST test

**Status**: Accepted

**Context**: House practice, unwritten until now: a Go file holds either
exported or unexported functions and methods, never both, so the public API
stays in a small file and implementation detail stays out of the reader's way
(attention dilution). An AST inventory showed production code already compliant
everywhere except flags.go (written this session) and 17 test files carrying
unexported helpers, 6 of them added this session.

**Decision**: Enforce the public/private file split with an AST test in
internal/layout

**Rationale**: Wrote the rule into CONVENTIONS.md and added
internal/layout/public_private_test.go, which parses every Go file under the
module root and fails on any file that mixes, in the same spirit as the route
packages' guard tests. Moved readFailure to flags_impl.go following the existing
*_impl.go pairing; moved all unexported test helpers into each package's
test_helper_test.go (testing_helper_test.go in sqlite/persist), creating helper
files where none existed. Chosen over a golangci custom linter (no plugin
infrastructure in the repo) and over leaving the rule as folklore.

**Consequence**: A mixed file fails make test with a message naming the
functions to move. Helper files are _test.go files (idiomatic Go, excluded from
binaries); the older non-_test test_helper.go pattern was retired in the same
change on the maintainer's decision.

---

## [2026-09-20-083424] Remove the dead CI harness ci/test and hack/test/test.sh

**Status**: Accepted

**Context**: ci/test/main.go drove 'spike initialization', 'spike login', 'spike
put' and 'spike get', none of which exist in the Pilot; it hard-coded
/home/volkan/Desktop/..., used the std logger, and was built only by
hack/test/test.sh, which nothing referenced and which called a
hack/test/cipher.sh that does not exist.

**Decision**: Remove the dead CI harness ci/test and hack/test/test.sh

**Rationale**: Dead code that cannot run is a broken window. Removal approved by
the user over rewriting against current commands. go mod tidy then drops
github.com/google/goexpect, the harness's only consumer.

**Consequence**: One fewer dependency; hack/test/ is gone. Live CLI coverage
remains with the opt-in integration tests (specs/integration-tests.md) and make
start checks.

---

## [2026-09-20-083424] Pilot exits non-zero; SDK logger routed to a file

**Status**: Accepted

**Context**: All 15 Pilot handlers printed errors and returned, so every failure
exited 0. The entry point used os.Exit(1), which CLAUDE.md forbids in favor of
log.FatalLn, but the SDK logger binds to os.Stdout on first use with no
configurable writer, so a fatal would put JSON on the data stream #299 had
cleaned. Asked, the user said: write to a well-known file; a stream not provided
does not mean we cannot use a file as a fallback for an exceptional case.

**Decision**: SPIKE Pilot exits non-zero on failure and routes the SDK logger to
a file

**Rationale**: Handlers are RunE and return errors; Execute prints 'Error:
<msg>' to stderr once (root has SilenceErrors/SilenceUsage) and terminates
through log.FatalLn. main first calls logfile.Route(), which opens
$HOME/.spike/pilot.log (or /tmp/.spike-$USER/pilot.log; 0600 in 0700, mirroring
the SDK recovery-dir resolution) and initializes the SDK logger while os.Stdout
temporarily points at it. Chosen over a documented os.Exit exception (keeps a
rule violation alive) and over accepting JSON on stdout (corrupts redirected
output). Partial-success paths (malformed key=value in secret put, missing
key/secret on get, failed cipher file close) became failures.
stdout.HandleAPIError(bool) became APIError(error); flags helpers return errors.

**Consequence**: Scripts can rely on the exit code. No os.Exit(1) remains in the
tree. The Route trick depends on the SDK's lazy binding; TASKS.md carries an
upstream item to add log.SetOutput to spike-sdk-go so the trick can go. Spec:
specs/pilot-failure-semantics.md.

---

## [2026-09-20-080852] Harden the lint gate: no presets, strict errcheck

**Status**: Accepted

**Context**: The v1-parity lint config kept default exclusion presets,
errcheck's built-in ignore list, and no security linter. Measured with those
removed on 2026-09-20: 396 findings (344 errcheck, 34 gosec, 18 staticcheck),
including silent strconv drops in the sqlite secret loader, ignored pflag
errors, unchecked Close/Remove on files holding shards, and unchecked
int/uint64/byte conversions on shard indices. The user rejected parity: SPIKE
handles secrets, so unhandled errors and unassessed suppressions are broken
windows.

**Decision**: Harden the lint gate: no presets, strict errcheck, staticcheck
all, gosec

**Rationale**: Config now: version 2, gosec enabled, errcheck with
disable-default-exclusions, check-blank and check-type-assertions, staticcheck
checks [all], no exclusion presets, no path exclusions. Every finding was fixed
at the site by four parallel workers on disjoint directories; the tree contains
zero nolint directives. gosec G101 tripped on three SQL statement constants
because their names contained Secret and the literal's leading characters passed
the entropy heuristic; rather than suppress, the constants were renamed after
what the statements do (QueryUpsertMetadata, QueryUpsertVersion,
QueryLoadMetadata), since a first nolint becomes the precedent every later one
cites. Fixing rather than excluding surfaced real defects: a broken benchmark
looking up a policy by an empty ID, a keeper ID that could wrap negative to a
huge Shamir index, a shard-distribution branch that did not zero the root
secret, key material printed in a keeper test, and a recovery shard hex-encoded
into an immutable string that could never be zeroed.

**Consequence**: make audit runs golangci-lint at this bar and reports 0 issues;
make test (race) passes with 357 tests in 23 packages. New helper package
app/spike/internal/cmd/flags fails fast on unreadable flags. The bar is no
nolint at all: a heuristic false positive is resolved by a rename or a
restructuring that keeps the name honest, not by a directive. Spec:
specs/lint-hardening.md.

---

## [2026-09-19-141657] Remediate grpc advisories: grpc v1.83.2, x/crypto 0.56

**Status**: Accepted

**Context**: govulncheck on the Go 1.27.1 bump reported two called advisories in
google.golang.org/grpc v1.79.3: GO-2026-6348 (fixed 1.83.1) and GO-2026-6061
(fixed 1.82.1). Output was byte-identical on base 42547a7e with the same
toolchain, so the findings were baseline. The user then authorized dependency
updates for these findings, excluding a blanket upgrade.

**Decision**: Remediate grpc advisories with targeted bumps: grpc v1.83.2,
x/crypto v0.56.0

**Rationale**: Fixed versions were verified against vuln.go.dev, not the scanner
summary. Chose grpc v1.83.2 over the minimal v1.83.1 because GO-2026-6443
(uncalled, fixed 1.83.2) sits in the same module and line; v1.84.0 was not
needed. Bumped x/crypto to v0.56.0 to clear three uncalled x/crypto/ssh
advisories with fixed versions (GO-2026-6303, -6354, -6355), following the Round
2 bar in specs/vuln-remediation.md. go-spiffe v2.6.0 -> v2.7.0 was not a choice:
grpc-go's go.mod requires it. x/net, x/sys, x/term, x/text, x/oauth2 and
genproto moved via go mod tidy.

**Consequence**: govulncheck: 0 called, 0 package-level; only GO-2026-5932
(openpgp, Fixed in N/A) remains. make test (race) and make audit exit 0 on the
new graph. No application-code changes. Recorded as Round 3 in
specs/vuln-remediation.md.

---

## [2026-09-19-140811] Run golangci-lint v2 in make audit to support Go 1.27

**Status**: Accepted

**Context**: Bumping go.mod to 1.27.1 makes CI (go-version-file: go.mod) install
Go 1.27.1. golangci-lint v1.64.8, the last v1 release, cannot read Go 1.27
export data (version 4 > max 2) and fails typecheck on every package, so the
audit gate would go red.

**Decision**: Run golangci-lint v2 in make audit to support Go 1.27

**Rationale**: Switched the Makefile step to
github.com/golangci/golangci-lint/v2 and migrated .golangci.yml to the v2
format. The v2 default linter set equals the v1 list (gosimple folded into
staticcheck); v1 default exclusions kept via presets; the new QF quick-fix
checks disabled so the lint scope is unchanged. Chosen over fixing the 7 QF
findings in a version-bump PR and over dropping golangci-lint.

**Consequence**: make audit step 7 passes on Go 1.27.1 with 0 issues. QF checks
can be enabled later as a separate change. Builder images bumped to
golang:1.27.1 for the same reason as #303.

---

## [2026-07-25-194549] Vulnerability bar is zero CALLED findings, not zero total

**Status**: Accepted

**Context**: The 2026-06-13 remediation set the bar at 'zero vulnerabilities
total, not merely zero called', clearing uncalled advisories too. GO-2026-5932
(golang.org/x/crypto/openpgp is unmaintained and unsafe by design) now makes
that unreachable: it reports 'Fixed in: N/A' because the package is deprecated
rather than patched, and it arrives transitively.

**Decision**: Vulnerability bar is zero CALLED findings, not zero total

**Rationale**: A criterion that cannot be met stops being a standard and starts
being noise that gets waived by habit. Amending it explicitly keeps the gate
meaningful. SPIKE does not call openpgp, so govulncheck exits 0 and CI passes
with the finding present; vendoring or forking x/crypto to excise it costs more
than the risk it removes.

**Consequence**: The standing bar is zero CALLED vulnerabilities, plus a
recorded justification in specs/vuln-remediation.md for every uncalled finding
left in place. Clearing uncalled findings is still preferred wherever a fixed
version exists. Revisit if an upstream stops depending on x/crypto/openpgp, or
if the finding ever becomes reachable.

---

## [2026-07-25-133218] Reserve spike/system/* namespaces against substring-matching policy patterns

**Status**: Accepted

**Context**: A responsible disclosure reported unanchored policy regexes as an
over-grant vulnerability. Substring matching by an unanchored regex is
documented, intended behavior, but policy management is gated by
CheckPolicyAccess against the literal path spike/system/acl, so a PathPattern of
'spike', 'system', or 'acl' matched that gate by substring and conferred control
over every policy in the system.

**Decision**: Reserve spike/system/* namespaces against substring-matching
policy patterns

**Rationale**: Implicit anchoring (the reporter's proposal) was rejected: it
patched only UpsertPolicy while sqlite/persist/regex.go recompiles patterns on
every load, so it fixed the memory backend and left SQLite exposed; and wrapping
in ^(?:...)$ silently converts working ^-only prefix policies into denials.
Instead the three reserved paths now require that a pattern DESCRIBE the path
(its full-match form ^(?:p)$ still matches) rather than merely contain it. A
purely syntactic ^...$ rule was implemented first and rejected because .* and
^.*$ are the same regex.

**Consequence**: Enforced in both UpsertPolicy (authoring-time rejection) and
CheckPolicyAccess (covers pre-existing stored policies and backend
recompilation). Ordinary paths keep plain substring semantics. Policies that
reached a reserved path via an unanchored pattern now fail loudly. See ADR-0033
and specs/policy-pattern-anchoring.md.

---

## [2026-07-18-110741] Bare-metal harness invokes SPIKE binaries via PATH, deliberately

**Status**: Accepted

**Context**: During the preflight work (2026-07-16) explicit-path invocation was
proposed to eliminate name-collision risk with the generic binary names (spike,
keeper, demo) and rejected; the rationale was never recorded.

**Decision**: Bare-metal harness invokes SPIKE binaries via PATH, deliberately

**Rationale**: Binaries on PATH are the user-facing convenience, and the harness
sharing that resolution forces PATH setup early, keeping one consistent story.
The preflight makes collisions loud through shadowing detection instead of
eliminating them.

**Consequence**: Do not re-propose explicit-path or prefixed binaries for the
dev harness; extend the preflight if new failure modes appear.

---

## [2026-07-17-080305] Config accessors crash fast on missing critical configuration

**Status**: Accepted

**Context**: A jira-era task proposed refactoring the env accessors (KeepersVal
and friends) to return sentinel errors instead of calling log.FatalLn; on
2026-07-17 a full SDK brief was drafted for it and withdrawn the same day.

**Decision**: Config accessors crash fast on missing critical configuration

**Rationale**: The crash is intentional: without critical configuration such as
SPIKE_NEXUS_KEEPER_PEERS, SPIKE cannot operate reliably, and failing fast beats
limping along misconfigured. The accessors are testable through the
SPIKE_STACK_TRACES_ON_LOG_FATAL panic-recover pattern, and the env-to-log
circular dependency the original task cited dissolved when both packages moved
into spike-sdk-go.

**Consequence**: Do not propose returned-error refactors for critical-config
accessors in spike or spike-sdk-go. New accessors for must-have configuration
should follow the same crash-fast idiom, keeping the panic-mode escape hatch for
tests.

---

## [2026-06-13-125427] Pin Go toolchain to 1.26.4 and bump circl/go-jose/x/net to clear govulncheck

**Status**: Accepted

**Context**: make audit (the pre-commit gate) failed: govulncheck reported 10
called vulnerabilities. 7 were Go 1.26.2 stdlib advisories
(textproto/mime/x509/html-template/net/net-http) and 3 were modules: x/net
v0.48.0, go-jose/v4 v4.1.3, circl v1.6.2. Pre-existing on main; unrelated to the
ctx/docs work in this branch.

**Decision**: Pin Go toolchain to 1.26.4 and bump circl/go-jose/x/net to clear
govulncheck

**Rationale**: Added 'toolchain go1.26.4' to go.mod (keeping the go 1.25.5
language baseline) so builds use the patched stdlib, and bumped circl->v1.6.3,
go-jose/v4->v4.1.4, x/net->v0.55.0, then go mod tidy. Chosen over (a) bumping
the go language directive to 1.26.4 (broader semantic change, unnecessary for
the CVEs) and (b) deferring remediation (leaves the audit gate red). govulncheck
gates on CALLED vulns only, so this clears the gate; uncalled import/module
advisories remain and resolve as deps bump over time.

**Consequence**: make audit is green (0 called vulnerabilities). Contributors
auto-download go1.26.4 via the toolchain directive. Transitive bumps to
x/crypto, x/sys, x/term, x/text. See also: specs/vuln-remediation.md

---

## [2026-06-13-121952] Own the ctx Makefile fragment as makefiles/Ctx.mk

**Status**: Accepted

**Context**: ctx init generates Makefile.ctx at the repo root and
regenerates/owns it, but SPIKE's convention places all make includes under
makefiles/*.mk (PascalCase: Main.mk, Test.mk). The generated root file violated
that convention.

**Decision**: Own the ctx Makefile fragment as makefiles/Ctx.mk

**Rationale**: Move the content to a project-owned makefiles/Ctx.mk and -include
it from the root Makefile, then gitignore the root Makefile.ctx so any
regenerated stray is neither included (no duplicate targets) nor committed.
Chosen over the default ctx pattern (include the generated root Makefile.ctx
directly), which keeps free upstream auto-updates but breaks the makefiles/*.mk
convention and clutters the repo root. For a small, rarely-changing fragment,
convention alignment and a single authoritative location win.

**Consequence**: makefiles/Ctx.mk is now project-owned and convention-aligned;
the root stays clean. Trade-off: it no longer auto-tracks upstream ctx changes
and must be manually reconciled if ctx updates its targets. See also:
specs/introduce-ctx.md
