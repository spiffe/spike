# Spec: Guard Reserved System Paths Against Unanchored Policy Patterns

## Problem Statement

SPIKE policy `SPIFFEIDPattern` and `PathPattern` fields are regular
expressions, matched with `regexp.MatchString`. That is substring
matching: a pattern with no `^`/`$` anchors matches anywhere inside the
candidate string. This is documented, intended behavior and it is how
Go's `regexp` package works.

A responsible disclosure (kanywst, 2026-07) reported this as an
over-grant vulnerability. The reported behavior reproduces, but the
diagnosis is imprecise: unanchored regexes matching substrings is not a
defect. The genuine defects the report surfaces are:

1. **Privilege escalation through reserved system paths.** Policy
   management is authorized by `CheckPolicyAccess(peer,
   "spike/system/acl", [write])`. A policy whose `PathPattern` is
   `spike`, `system`, or `acl` therefore authorizes that workload to
   create and modify arbitrary policies, including granting itself
   `super`. `PathPattern: "spike"` is plausible for an operator who
   keeps secrets under a `spike/` namespace. The same applies to
   `spike/system/secret` and `spike/system/cipher/exec`. Nothing in
   the codebase reserves these namespaces.

   The identity half is equally severe: a policy legitimately scoped to
   `spike/system/acl` but whose `SPIFFEIDPattern` is unanchored (for
   example `spiffe://example\.org/admin`) also matches
   `spiffe://example.org/admin-attacker`.

2. **Documentation that contradicts itself.** 88% of the 147 concrete
   pattern examples in the repository are anchored, and three documents
   teach anchoring correctly. But `usage/commands/policy.md` carries a
   block titled "Path Pattern Examples" whose entries are `^`-only, with
   an inline comment (`^secrets/database/creds # Only the specific creds
   resource`) that is factually false. Its "Common Errors" section hands
   the reader unanchored patterns as the correct remediation.
   `CLAUDE.md` frames the only correctness axis as regex-versus-glob and
   marks two unanchored patterns as correct.
   `examples/federation/workload-set-policies.sh` ships a literal glob
   with `write` permission.

## Rejected Solution

The reporter proposed wrapping both patterns at compile time in
`UpsertPolicy`:

```go
idRegex, _ := regexp.Compile("^(?:" + policy.SPIFFEIDPattern + ")$")
```

This is rejected on two grounds, both verified empirically:

- **It does not fix production.** `UpsertPolicy` is not the only compile
  site. `app/nexus/internal/state/backend/sqlite/persist/regex.go`
  recompiles both patterns from the stored strings on every load, and
  `CheckPolicyAccess` reaches that path through `ListPolicies` ->
  `LoadAllPolicies` on every access check. With the patch applied, the
  reporter's test passes under the memory backend (which retains the
  struct built by `UpsertPolicy`) while the identical scenario under
  SQLite still over-grants on both path and identity.
- **It silently revokes legitimate access.** A prefix policy written
  `^tenants/acme/` matches `tenants/acme/db/creds` today. Wrapped, it
  becomes `^(?:^tenants/acme/)$` and matches nothing usable. The
  repository ships eight such `^`-only patterns. Anchoring on upgrade
  turns them into silent denials with no error naming the cause.

Blanket anchoring also changes `PathPattern` from "a regular expression"
into "a regular expression that is implicitly full-match", a different
dialect than what is documented and what operators were told to expect.

## Proposed Solution

Keep regex semantics exactly as they are. Constrain only the case where
an accident is catastrophic, and make the constraint loud rather than
silent.

1. **Reserve the system namespaces.** Introduce a guard over
   `spike/system/acl`, `spike/system/secret`, and
   `spike/system/cipher/exec`. A policy may reach a reserved system path
   only when it **describes** that path rather than merely appearing
   inside it.

   The test is semantic, not syntactic: a pattern describes a path when
   its full-match form, `^(?:pattern)$`, still matches the path. So
   `^spike/system/acl$`, `spike/system/acl`, `^spike/system/.*$`, and
   `.*` all qualify, because each of their authors plainly meant to
   include the reserved path. `acl`, `system`, and `spike` do not: they
   reach it only because Go regular expressions match substrings, and
   read as descriptions of a path they mean something else entirely.

   A purely syntactic "must start with `^` and end with `$`" rule was
   implemented first and rejected. `.*` and `^.*$` are the same regular
   expression, so accepting one and refusing the other polices spelling
   without changing what is granted, and it broke the wildcard policies
   already in the test suite.

   The identity side is held to a related rule. An unanchored
   `SPIFFEIDPattern` is a substring matcher, so a policy written for
   `spiffe://example\.org/admin` also admits
   `spiffe://example.org/admin-attacker`. For reserved paths the pattern
   must therefore be anchored, or be an unambiguous catch-all such as
   `.*` (detected by testing the full-match form against a probe value
   no real SPIFFE ID would equal). Narrowing a catch-all remains the
   operator's call.

2. **Enforce at both ends, not just at creation.** `UpsertPolicy`
   rejects a violating policy with an actionable error, so the operator
   learns at authoring time. `CheckPolicyAccess` independently skips a
   policy that reaches a reserved path without anchors, so policies
   already stored before this change cannot escalate either. This
   deliberately avoids the single-compile-site mistake that makes the
   rejected patch ineffective.

3. **Fix the documentation.** Correct every unanchored and partially
   anchored example that is presented as usable, correct the false
   comment, and add an explicit, prominent statement that patterns are
   regular expressions matched as substrings, that supplying anchors is
   the operator's responsibility, and that patterns should be narrowed
   to the smallest set that satisfies the need.

## File Surface

Code:

- `app/nexus/internal/state/base/reserved.go` (new): reserved path list,
  anchor predicate, and the guard used by both call sites.
- `app/nexus/internal/state/base/reserved_test.go` (new): escalation
  regression tests, including a SQLite-backed test proving the guard
  survives the recompile-on-load path.
- `app/nexus/internal/state/base/policy.go`: call the guard from
  `UpsertPolicy` and from `CheckPolicyAccess`.

Documentation:

- `docs-src/content/usage/commands/policy.md`: rewrite "Pattern Syntax",
  fix the "Path Pattern Examples" and "SPIFFE ID Pattern Examples"
  blocks, fix the two unanchored "Common Errors" remediations, fix the
  `admin-policy` and `admin-service` examples, document the reserved
  namespace rule, extend "Best Practices".
- `CLAUDE.md`: add anchoring to the pattern convention; fix the two
  unanchored examples and the leading-slash contradiction.
- `docs-src/content/architecture/adrs/adr-0030.md`: anchor the sample
  audit-log pattern.
- `docs-src/content/recipes/writing-access-policies.md`,
  `granting-a-workload-access.md`: add the reserved namespace note.
- `app/nexus/internal/route/acl/policy/get.go`,
  `app/spike/internal/cmd/policy/apply.go`,
  `app/spike/internal/cmd/policy/get.go`: fix doc-comment examples.
- `examples/federation/workload-set-policies.sh`: replace the glob.
- `examples/policies/sample-policy.yaml`: fix the wrong YAML keys.
- `diagrams/011-spike-pilot-policy-creation.md`: note anchoring.

## Error / Edge Cases

- **An operator legitimately delegating policy management** writes
  `^spike/system/acl$`, which the guard accepts. The previously
  documented unanchored spelling now fails with an error naming the
  anchored form. This is a visible, actionable break, not a silent one.
- **Wildcard administrative policies** (`.*` and `^.*$` alike) remain
  accepted and behave exactly as before. They describe every path,
  including the reserved ones, and say so unambiguously.
- **Policies stored before this change** are not rewritten or migrated.
  `CheckPolicyAccess` refuses to honor them for reserved paths, and the
  refusal is logged so an operator can find and fix them.
- **A pattern that fails to compile** inside the full-match wrapper is
  treated as not describing the path, so it is refused rather than
  admitted. `UpsertPolicy` has already rejected uncompilable patterns by
  that point, so this only matters as a safe default.
- **Non-system paths are entirely unaffected.** Substring matching still
  applies everywhere else, exactly as documented.

## Non-Goals

- Not changing how `PathPattern` or `SPIFFEIDPattern` are compiled for
  ordinary paths. They stay plain, unanchored Go regular expressions.
- Not adding an environment variable to opt out of the guard.
- Not migrating or rewriting stored policies.
- Not adding a create-time warning for unanchored non-system patterns.
  Considered, deferred: it belongs with a broader policy-lint feature
  rather than this security fix.

## Verification

- New tests in `reserved_test.go` fail against the pre-fix tree and pass
  after, under both the memory and SQLite backends.
- The reporter's original reproduction still demonstrates substring
  matching for ordinary paths, which is intended and documented.
- `make lint` and `make test` pass across the whole project.
