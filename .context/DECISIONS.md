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
