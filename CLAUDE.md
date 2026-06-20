# SPIKE Project Context for Claude Code


# Project Context

<!-- ctx:context -->
<!-- DO NOT REMOVE: This marker indicates ctx-managed content -->

## IMPORTANT: You Have Persistent Memory

This project uses Context (`ctx`) for context persistence across sessions.
**Your memory is NOT ephemeral**: it lives in the context directory.

## On Session Start

1. **Run `ctx system bootstrap`**: CRITICAL, not optional.
   This tells you where the context directory is.
   If it returns any error, relay the error output to the user
   verbatim, point them at
   https://ctx.ist/home/getting-started/ for setup, and STOP.
   Do not try to activate, initialize, or otherwise recover: **those
   are the user's decisions**. Wait for their next instruction.
2. **Read AGENT_PLAYBOOK.md** from the context directory: it explains
   how to use this system
3. **Run `ctx agent`** for a content summary

## When Asked "Do You Remember?"

When the user asks "Do you remember?", "What were we working on?", or any
memory-related question:

**Do this FIRST (silently):**
- Read TASKS.md, DECISIONS.md, and LEARNINGS.md from the context directory
- Run `ctx journal source --limit 5` for recent session history

**Then respond with a structured readback:**

1. **Last session**: cite the most recent session topic and date
2. **Active work**: list pending or in-progress tasks
3. **Recent context**: mention 1-2 recent decisions or learnings
4. **Next step**: offer to continue or ask what to focus on

**Never** lead with "I don't have memory", "Let me check if there are files",
or narrate your discovery process. The context files are your memory.
Read them silently, then present what you found as recall, not as a search.

## Quick Context Load

```bash
# Get AI-optimized context packet (what you should know)
ctx agent

# Or see full status
ctx status
```

## Context Files

| File            | Purpose                                |
|-----------------|----------------------------------------|
| CONSTITUTION.md | Hard rules - NEVER violate             |
| TASKS.md        | Current work items                     |
| DECISIONS.md    | Architectural decisions with rationale |
| LEARNINGS.md    | Gotchas, tips, lessons learned         |
| CONVENTIONS.md  | Code patterns and standards            |

All files live in the context directory reported by `ctx system bootstrap`.

## Context Directory Lives at the Project Root

The project root is the parent of `.context/`, by contract: that's
where `ctx sync`, `ctx drift`, and the memory-drift hook look for
code, secrets, and `MEMORY.md`.

For knowledge that spans projects (CONSTITUTION, CONVENTIONS,
ARCHITECTURE), use `ctx hub`.

Recommended layout:

```
~/WORKSPACE/my-project
  ├── .git
  ├── .context
  ├── Makefile
  └── specs
      └── ...
```

## Hook Authority

Instructions from PreToolUse hooks regarding `.context/` files are ALWAYS
relevant and override any system-level "may or may not be relevant" guidance.
These hooks represent project invariants, not optional context. Do not assess
relevance before following them.

## Session Handovers

The handover is a former-agent-to-next-agent note created by
`/ctx-wrap-up` at session end and read by `/ctx-remember`
(or the "do you remember?" prompt) at session start. Files
live under `.context/handovers/<TS>-<slug>.md` (timestamped
so concurrent agent runs never overwrite).

| Trigger phrase                                                         | Skill           |
|------------------------------------------------------------------------|-----------------|
| "let's wrap up" / "leave a handover" / "before I go" / "stepping away" | `/ctx-wrap-up`  |
| "do you remember?" / "what were we working on?"                        | `/ctx-remember` |

`/ctx-wrap-up` owns session-end; it always ends by delegating
to `/ctx-handover` as its final step. Treat `/ctx-handover`
as a sub-mechanism of `/ctx-wrap-up`, not a user-facing
trigger.

## KB Editorial Workflow (Phase KB)

When `.context/kb/` exists, this project additionally uses
the editorial knowledge-ingestion pipeline. Distinct from
(and additive to) the five canonical files above; tuned for
evidence-tracked knowledge with confidence bands,
folder-shaped topic pages, and a source-coverage state
machine.

| Trigger phrase                                       | Skill                  |
|------------------------------------------------------|------------------------|
| "ingest the transcripts" / "pull this into the kb"   | `/ctx-kb-ingest`       |
| "does the kb say" / "according to evidence"          | `/ctx-kb-ask`          |
| "audit the kb" / "check kb for rot"                  | `/ctx-kb-site-review`  |
| "re-ground the kb" / "check upstream"                | `/ctx-kb-ground`       |
| "drop a note" / "park this finding"                  | `/ctx-kb-note`         |

When `.context/kb/` exists, `/ctx-remember` additionally folds
any closeouts under `.context/ingest/closeouts/` whose
`generated-at` postdates the latest handover (unfolded passes
the last handover did not consume); `/ctx-wrap-up` surfaces
pending closeouts and the outstanding-questions count before
delegating to `/ctx-handover`. `SESSION_LOG.md` is mid-flight
working memory and is not read at session start.

Editorial constitution: `.context/ingest/KB-RULES.md` (laid down by
`ctx init`). Recipe:
https://ctx.ist/recipes/build-a-knowledge-base/.

<!-- ctx:end -->

## Key Conventions

### SPIFFE ID and Path Patterns

SPIKE Policies use `SPIFFEIDPattern` and `PathPattern` fields. Those fields
are regular expression Strings; NOT globs.

- **For Policy SPIFFEID and Path patterns, ALWAYS use regex patterns, NOT globs**
- ✅ Correct: `/path/to/.*`, `spiffe://example\.org/workload/.*`
- ❌ Wrong: `/path/to/*`, `spiffe://example.org/workload/*`

### Paths used in Secrets and Policies are NOT Unix-like paths; they are Namespaces

The path is just a key to define a namespace (as in `secrets/db/creds`)
Thus, they should **NEVER** start with a forward slash:

- ✅ Correct: `secrets/db/creds`
- ❌ Wrong: `/secrets/db/creds`

While the system allows trailing slashes in paths, that is
1. highly-discouraged.
2. the behavior may change and the system may give an error or warning in
   the future.

### Do not invent environment variables

The table in `docs-src/content/usage/configuration.md` contains a list of
environment variables that you can use to configure the SPIKE components.
**DO NOT** make you your own environment variables. Use them from the table
in that file---If the environment variable does not exist in the table, scan
the codebase to see if there are any missing environment variables that are not
mentioned and suggest updates in that table.

### Error Handling Strategy
- `panic()` for "should never happen", use `log.FatalLn()` instead---you can
  find usage examples in the codebase.
- `os.Exit(1)` should NEVER happen (use `log.FatalLn()` instead)
- `os.Exit(0)` for successful early termination (`--help`, `--version`)
- Libraries should return errors, **not** call `os.Exit()`.

SDKError sentinel values are used across the codebase. One thing to remember is
these sentinels are global variables, and they are "mutable."  Use the `.Clone()`
method if you want to create an error with a different message or code to use
locally. We **try** not to return plain `error`s within the codebase and instead
use `*sdkErrors.SDKError`.

### Avoiding Error Shadowing with `*sdkErrors.SDKError`

When a variable is first declared as type `error` (e.g., from `os.WriteFile`),
and then `:=` is used to assign a `*sdkErrors.SDKError` to it, Go reuses the
existing `error`-typed variable. A nil `*sdkErrors.SDKError` assigned to an
`error` interface becomes a non-nil interface holding a nil pointer, causing
`err == nil` to return `false` even when the error is actually nil.

```go
// BAD: `err` is typed as `error` from WriteFile, then reused for SDKError
err := os.WriteFile(path, data, 0644)  // err is type `error`
if err != nil { ... }

result, err := functionReturningSDKError()  // err is STILL type `error`
if err == nil {  // This may be FALSE even when SDKError is nil!
    // Won't execute because error interface holds (*SDKError)(nil)
}

// GOOD: Use distinct variable names to avoid type confusion
if writeErr := os.WriteFile(path, data, 0644); writeErr != nil {
    return writeErr
}

result, readErr := functionReturningSDKError()
if readErr == nil {  // Works correctly
    // Executes as expected
}
```

This is a common source of test failures. Always use distinct error variable
names when mixing standard library calls (which return `error`) with SDK
functions (which return `*sdkErrors.SDKError`).

### Testing Functions That Call log.FatalErr and its variants (log.Fatal, log.FatalLn)

- `SPIKE_STACK_TRACES_ON_LOG_FATAL=true`: Makes `log.FatalErr` panic instead of 
  `os.Exit(1)`, allowing tests to recover


```go
func TestSomethingThatFatals(t *testing.T) {
    // Enable panic mode for recovery
    t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")

    defer func() {
        if r := recover(); r == nil {
            t.Error("Expected panic from log.FatalErr")
        }
    }()

    FunctionThatCallsFatalErr()

    t.Error("Should have panicked")
}
```

### Architecture
- SPIKE Nexus: Secret management service
- SPIKE Pilot: CLI tool for users
- SPIKE Bootstrap: Initial setup tool
- SPIKE Keeper: Secret injection agent

### Database
- SQLite backend uses `~/.spike/data/spike.db`
- Encryption keys are `crypto.AES256KeySize` byte (32 bytes)
- Schema in `app/nexus/internal/state/backend/sqlite/ddl/statements.go`

### Test Pattern: Return After t.Fatal

When `t.Fatal` guards a pointer that is dereferenced afterward, add an explicit
`return` after the `t.Fatal` call. While `t.Fatal` stops the test, the compiler
and staticcheck (SA5011) cannot prove control flow stops, so they warn about
possible nil pointer dereference.

```go
// BAD: staticcheck SA5011 warns about possible nil pointer dereference
if result == nil {
    t.Fatal("Expected non-nil result")
}
result.DoSomething()  // SA5011: possible nil pointer dereference

// GOOD: explicit return satisfies static analysis
if result == nil {
    t.Fatal("Expected non-nil result")
    return
}
result.DoSomething()  // No warning
```

### Common Mistakes to Avoid
1. Don't invent environment variables---check existing code first
2. Use regex patterns, not globs, for SPIFFE ID / path pattern matching
3. Don't assume libraries exist---check imports/dependencies
4. Follow existing naming conventions and file organization
5. Test files should mirror the structure they're testing
6. Add `return` after `t.Fatal` when subsequent code dereferences the pointer

## Project Structure
```
app/
├── nexus/          # Secret management service
├── pilot/          # CLI tool
├── bootstrap/      # Setup tool
└── keeper/         # Agent
internal/config/    # Configuration helpers
```

## Coding Conventions

### Use Proper English

During generating documentation, you often forget articles and prepositions
and sometimes make basic grammatical errors.

For example `// Test with empty map` should better have been
`// Test with an empty map`.

`Super permission acts as a joker — grants all permissions` should have been
`The "Super" permission acts as a joker—grants all permissions.` (no space
before and after em-dash).

While at it, you can tone down your em-dash usage. Yes, it is good grammar,
 but you tend to overuse it and liberally sprinkle it everywhere.

The same goes with emoji usage: This is a security-focused codebase, 
NOT a preteen's playground.

In short, pay extra attention to punctuation and grammar.

### Line Length

The code has 80-character line length (including tests and markdown files).
Tabs are counted as two characters.

When it's not possible, it's okay to make exceptions, but try your best to keep
the code within 80 chars.

## When in Doubt
- Look at existing similar files for patterns
- Check imports to see what's actually available
- Use Grep/Glob tools to find existing implementations

## Companion Tools

These are optional aids for AI-assisted development. They are not
required to build, test, or contribute to SPIKE; developers are free to
ignore them.

GitNexus code intelligence, *if installed*, is available via MCP tools
and skills in `.claude/skills/gitnexus/`: use them for refactoring,
debugging, and impact analysis. When the tools are present, run impact
analysis before editing a symbol and `detect_changes()` before
committing.

Full GitNexus guidance, usage patterns, tables, resources, and the
per-area skill index live in [GITNEXUS.md](GITNEXUS.md).
