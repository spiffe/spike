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
## [2026-07-25-133218] Reserve spike/system/* namespaces against substring-matching policy patterns

**Status**: Accepted

**Context**: A responsible disclosure reported unanchored policy regexes as an over-grant vulnerability. Substring matching by an unanchored regex is documented, intended behavior, but policy management is gated by CheckPolicyAccess against the literal path spike/system/acl, so a PathPattern of 'spike', 'system', or 'acl' matched that gate by substring and conferred control over every policy in the system.

**Decision**: Reserve spike/system/* namespaces against substring-matching policy patterns

**Rationale**: Implicit anchoring (the reporter's proposal) was rejected: it patched only UpsertPolicy while sqlite/persist/regex.go recompiles patterns on every load, so it fixed the memory backend and left SQLite exposed; and wrapping in ^(?:...)$ silently converts working ^-only prefix policies into denials. Instead the three reserved paths now require that a pattern DESCRIBE the path (its full-match form ^(?:p)$ still matches) rather than merely contain it. A purely syntactic ^...$ rule was implemented first and rejected because .* and ^.*$ are the same regex.

**Consequence**: Enforced in both UpsertPolicy (authoring-time rejection) and CheckPolicyAccess (covers pre-existing stored policies and backend recompilation). Ordinary paths keep plain substring semantics. Policies that reached a reserved path via an unanchored pattern now fail loudly. See ADR-0033 and specs/policy-pattern-anchoring.md.

---

## [2026-07-18-110741] Bare-metal harness invokes SPIKE binaries via PATH, deliberately

**Status**: Accepted

**Context**: During the preflight work (2026-07-16) explicit-path invocation was proposed to eliminate name-collision risk with the generic binary names (spike, keeper, demo) and rejected; the rationale was never recorded.

**Decision**: Bare-metal harness invokes SPIKE binaries via PATH, deliberately

**Rationale**: Binaries on PATH are the user-facing convenience, and the harness sharing that resolution forces PATH setup early, keeping one consistent story. The preflight makes collisions loud through shadowing detection instead of eliminating them.

**Consequence**: Do not re-propose explicit-path or prefixed binaries for the dev harness; extend the preflight if new failure modes appear.

---

## [2026-07-17-080305] Config accessors crash fast on missing critical configuration

**Status**: Accepted

**Context**: A jira-era task proposed refactoring the env accessors (KeepersVal and friends) to return sentinel errors instead of calling log.FatalLn; on 2026-07-17 a full SDK brief was drafted for it and withdrawn the same day.

**Decision**: Config accessors crash fast on missing critical configuration

**Rationale**: The crash is intentional: without critical configuration such as SPIKE_NEXUS_KEEPER_PEERS, SPIKE cannot operate reliably, and failing fast beats limping along misconfigured. The accessors are testable through the SPIKE_STACK_TRACES_ON_LOG_FATAL panic-recover pattern, and the env-to-log circular dependency the original task cited dissolved when both packages moved into spike-sdk-go.

**Consequence**: Do not propose returned-error refactors for critical-config accessors in spike or spike-sdk-go. New accessors for must-have configuration should follow the same crash-fast idiom, keeping the panic-mode escape hatch for tests.

---

## [2026-06-13-125427] Pin Go toolchain to 1.26.4 and bump circl/go-jose/x/net to clear govulncheck

**Status**: Accepted

**Context**: make audit (the pre-commit gate) failed: govulncheck reported 10 called vulnerabilities. 7 were Go 1.26.2 stdlib advisories (textproto/mime/x509/html-template/net/net-http) and 3 were modules: x/net v0.48.0, go-jose/v4 v4.1.3, circl v1.6.2. Pre-existing on main; unrelated to the ctx/docs work in this branch.

**Decision**: Pin Go toolchain to 1.26.4 and bump circl/go-jose/x/net to clear govulncheck

**Rationale**: Added 'toolchain go1.26.4' to go.mod (keeping the go 1.25.5 language baseline) so builds use the patched stdlib, and bumped circl->v1.6.3, go-jose/v4->v4.1.4, x/net->v0.55.0, then go mod tidy. Chosen over (a) bumping the go language directive to 1.26.4 (broader semantic change, unnecessary for the CVEs) and (b) deferring remediation (leaves the audit gate red). govulncheck gates on CALLED vulns only, so this clears the gate; uncalled import/module advisories remain and resolve as deps bump over time.

**Consequence**: make audit is green (0 called vulnerabilities). Contributors auto-download go1.26.4 via the toolchain directive. Transitive bumps to x/crypto, x/sys, x/term, x/text. See also: specs/vuln-remediation.md

---

## [2026-06-13-121952] Own the ctx Makefile fragment as makefiles/Ctx.mk

**Status**: Accepted

**Context**: ctx init generates Makefile.ctx at the repo root and regenerates/owns it, but SPIKE's convention places all make includes under makefiles/*.mk (PascalCase: Main.mk, Test.mk). The generated root file violated that convention.

**Decision**: Own the ctx Makefile fragment as makefiles/Ctx.mk

**Rationale**: Move the content to a project-owned makefiles/Ctx.mk and -include it from the root Makefile, then gitignore the root Makefile.ctx so any regenerated stray is neither included (no duplicate targets) nor committed. Chosen over the default ctx pattern (include the generated root Makefile.ctx directly), which keeps free upstream auto-updates but breaks the makefiles/*.mk convention and clutters the repo root. For a small, rarely-changing fragment, convention alignment and a single authoritative location win.

**Consequence**: makefiles/Ctx.mk is now project-owned and convention-aligned; the root stays clean. Trade-off: it no longer auto-tracks upstream ctx changes and must be manually reconciled if ctx updates its targets. See also: specs/introduce-ctx.md
