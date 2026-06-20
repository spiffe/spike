# Agent Playbook

<!--
RELATED: AGENT_PLAYBOOK_GATE.md is a distilled subset of this file,
injected at session start by the context-load-gate hook. If you change
directives here, check whether the gate file needs a corresponding update.
-->

## Mental Model

Each session is a fresh execution in a shared workshop. Work
continuity comes from artifacts left on the bench. Follow the
cycle: **Work → Reflect → Persist**. After completing a task,
making a decision, learning something, or hitting a milestone:
persist before continuing. Don't wait for session end; it may
never come cleanly.

## File Interaction Protocol

When a task involves reading, modifying, or reasoning about a file:

1. **Read before act**
    - Read the file content directly before making any change
    - Do not rely on memory, summaries, or prior reads
2. **No partial reads**
    - Do not sample the beginning or end of a file and assume the rest
3. **Freshness requirement**
    - A read must be recent relative to the action
    - Do not reuse stale context from earlier in the session
4. **No implicit scope**
    - "This change is small" is not a valid justification
    - "This file is large" is not a valid justification
5. **Edit authority comes from visibility**
    - If you haven't seen it, you don't get to modify it

## Spec Requirement

Do not begin implementation work without a spec.

- Every implementation task must trace to a spec file
- **If no spec exists, STOP and create one first**
- Do not treat task text alone as a substitute for a spec

## Independent Review

Sub-agent review is not optional once implementation begins.

A review must be invoked when ANY of the following occur:

- Before the first modification to the codebase
- After completing one or more tasks in TASKS.md
- Before declaring the work complete

Required review inputs:
- the governing spec
- TASKS.md
- the current implementation

Review prompt:
- "Review <spec-file>, TASKS.md, and the current implementation for drift,
  omissions, invalid assumptions, and incomplete requirements."

Do not declare work complete until review findings are either resolved or
explicitly recorded.

## Invoking ctx

Always use `ctx` from PATH:
```bash
ctx status        # ✓ correct
ctx agent         # ✓ correct
./dist/ctx        # ✗ avoid hardcoded paths
go run ./cmd/ctx  # ✗ avoid unless developing ctx itself
```
Check with `which ctx` if unsure whether it's installed.

### When ctx Returns an Error

Triage the error before reacting:

- **Invocation error**: the message points at your call: unknown
  flag, unknown command, wrong argument count, missing required
  flag. Read `ctx <command> --help`, fix the call, and retry.
- **Everything else**: missing context directory, config problem,
  hook rejection, permission denied, unexpected failure. Relay the
  output to the user **verbatim** and stop. Do not add flags, run
  other commands, edit files to fix the cause, or retry. Wait for
  the user's next instruction.

When unsure which kind you're looking at, treat it as the second.

## Context Readback

Before starting any work, read the required context files and confirm to the
user: "I have read the required context files and I'm following project
conventions." Do not begin implementation until you have done so.

## Supplementary Files

These files live in `.context/` alongside the core context files.
Read them when the task at hand warrants it, not on every session.

| File               | Read when                                                      |
|--------------------|----------------------------------------------------------------|
| ARCHITECTURE.md    | Working on structure, adding packages, or tracing flow         |
| DETAILED_DESIGN.md | Deep-diving into internals (generated via `/ctx-architecture`) |
| GLOSSARY.md        | Encountering unfamiliar project-specific terminology           |

## Context Directory Lives at the Project Root

The project root is the parent of `.context/`, by contract —
specifically `filepath.Dir(ContextDir())`. That's where `ctx sync`,
`ctx drift`, and the memory-drift hook look for code, secrets,
and `MEMORY.md`.

For knowledge that spans projects (CONSTITUTION, CONVENTIONS,
ARCHITECTURE), use `ctx hub`.

Recommended layout:

```
~/WORKSPACE/my-project
  ├── .git
  ├── .context
  ├── Makefile
  ├── Makefile.ctx
  └── specs
      └── ...
```

## Reason Before Acting

Before implementing any non-trivial change, think through it step-by-step:

1. **Decompose**: break the problem into smaller parts
2. **Identify impact**: what files, tests, and behaviors does this touch?
3. **Anticipate failure**: what could go wrong? What are the edge cases?
4. **Sequence**: what order minimizes risk and maximizes checkpoints?

This applies to debugging too: reason through the cause before reaching
for a fix. Rushing to code before reasoning is the most common source of
wasted work.

### Chunk and Checkpoint Large Tasks

For work spanning many files or steps, break it into independently
verifiable chunks. After each chunk:

1. **Commit**: save progress to git so nothing is lost
2. **Persist**: record learnings or decisions discovered during the chunk
3. **Verify**: run tests or `make lint` before moving on

Track progress via TASKS.md checkboxes. If context runs low mid-task,
persist a progress note (what's done, what's next, what assumptions
remain) before continuing in a new window. The `check-context-size`
hook nudges at 60% usage (checkpoint) and warns at 90% (urgent):
treat these as signals to persist progress, not to rush.

## Session Lifecycle

A session follows this arc:

**Load → Orient → Pick → Work → Commit → Reflect**

Not every session uses every step: a quick bugfix skips reflection, a
research session skips committing: but the full flow is:

<!-- drift-check: ls internal/assets/claude/skills/ctx-remember internal/assets/claude/skills/ctx-status internal/assets/claude/skills/ctx-next internal/assets/claude/skills/ctx-implement internal/assets/claude/skills/ctx-commit internal/assets/claude/skills/ctx-reflect 2>&1 | grep -c skills/ -->
| Step        | What Happens                                       | Skill / Command  |
|-------------|----------------------------------------------------|------------------|
| **Load**    | Recall context, present structured readback        | `/ctx-remember`  |
| **Orient**  | Check context health, surface issues               | `/ctx-status`    |
| **Pick**    | Choose what to work on                             | `/ctx-next`      |
| **Work**    | Write code, fix bugs, research                     | `/ctx-implement` |
| **Commit**  | Commit with context capture                        | `/ctx-commit`    |
| **Reflect** | Surface persist-worthy items from this session     | `/ctx-reflect`   |

### Context Health at Session Start

During **Load** and **Orient**, run `ctx status` and read the output.
Surface problems worth mentioning:

- **High completion ratio in TASKS.md**: offer to archive
- **Stale context files** (not modified recently): mention before
  stale context influences work
- **Bloated token count** (over 30k): offer `ctx compact`
- **Drift between files and code**: spot-check paths from
  ARCHITECTURE.md against the actual file tree

One sentence is enough: don't turn startup into a maintenance session.

### Context Window Limits

The `check-context-size` hook (`ctx system check-context-size`) monitors
context window usage. It nudges at 60% (one-shot checkpoint) and warns
at 90% (recurring urgent). When you see either signal or sense context
is running long:

- **Persist progress**: write what's done and what's left to TASKS.md
  or a progress note
- **Checkpoint state**: commit work-in-progress so a fresh session can
  pick up cleanly
- **Summarize**: leave a breadcrumb for the next window: the current
  task, open questions, and next step

Context compaction happens automatically, but the next window loses
nuance. Explicit persistence is cheaper than re-discovery.

### Conversational Triggers

Users rarely invoke skills explicitly. Recognize natural language:

<!-- drift-check: ls internal/assets/claude/skills/ctx-remember internal/assets/claude/skills/ctx-status internal/assets/claude/skills/ctx-next internal/assets/claude/skills/ctx-commit internal/assets/claude/skills/ctx-reflect internal/assets/claude/skills/ctx-decision-add internal/assets/claude/skills/ctx-learning-add internal/assets/claude/skills/ctx-convention-add internal/assets/claude/skills/ctx-task-add 2>&1 | grep -c skills/ -->
| User Says                                       | Action                                                 |
|-------------------------------------------------|--------------------------------------------------------|
| "Do you remember?" / "What were we working on?" | `/ctx-remember`                                        |
| "How's our context looking?"                    | `/ctx-status`                                          |
| "What should we work on?"                       | `/ctx-next`                                            |
| "Commit this" / "Ship it"                       | `/ctx-commit`                                          |
| "The rate limiter is done" / "We finished that" | `ctx task complete` (match to TASKS.md)                |
| "What did we learn?"                            | `/ctx-reflect`                                         |
| "Save that as a decision"                       | `/ctx-decision-add`                                    |
| "That's worth remembering" / "Any gotchas?"     | `/ctx-learning-add`                                    |
| "Record that convention"                        | `/ctx-convention-add`                                  |
| "Add a task for that"                           | `/ctx-task-add`                                        |
| "Sync memory" / "What's in auto memory?"        | `ctx memory sync` / `ctx memory status`                |
| "Import from memory"                            | `ctx memory import --dry-run` then `ctx memory import` |
| "Let's wrap up"                                 | Reflect → persist outstanding items → present together |

## Proactive Persistence

**Don't wait to be asked.** Identify persist-worthy moments in real time:

| Event                                      | Action                                                            |
|--------------------------------------------|-------------------------------------------------------------------|
| Completed a task                           | Mark done in TASKS.md, offer to add learnings                     |
| Chose between design alternatives          | Offer: *"Worth recording as a decision?"*                         |
| Hit a subtle bug or gotcha                 | Offer: *"Want me to add this as a learning?"*                     |
| Finished a feature or fix                  | Identify follow-up work, offer to add as tasks                    |
| Resolved a tricky debugging session        | Capture root cause before moving on                               |
| Multi-step task or feature complete        | Suggest reflection: *"Want me to capture what we learned?"*       |
| Session winding down                       | Offer: *"Want me to capture outstanding learnings or decisions?"* |
| Shipped a feature or closed batch of tasks | Offer blog post or journal site rebuild                           |

**Self-check**: periodically ask yourself: *"If this session ended
right now, would the next session know what happened?"* If no, persist
something before continuing.

Offer once and respect "no." Default to surfacing the opportunity
rather than letting it pass silently.

### Task Lifecycle Timestamps

Track task progress with timestamps for session correlation:

```markdown
- [ ] Implement feature X #added:2026-01-25-220332
- [ ] Fix bug Y #added:2026-01-25-220332 #started:2026-01-25-221500
- [x] Refactor Z #added:2026-01-25-200000 #started:2026-01-25-210000
```

| Tag        | When to Add                              | Format               |
|------------|------------------------------------------|----------------------|
| `#added`   | Auto-added by `ctx task add`             | `YYYY-MM-DD-HHMMSS`  |
| `#started` | When you begin working on the task       | `YYYY-MM-DD-HHMMSS`  |

## Collaboration Defaults

Standing behavioral defaults for how the agent collaborates with the
user. These apply unless the user overrides them for the session
(e.g., "skip the alternatives, just build it").

- **At design decisions**: always present 2+ approaches with
  trade-offs before committing: don't silently pick one
- **At completion claims**: map claims to evidence (e.g., "tests
  pass" requires 0-failure output, "build succeeds" requires exit 0).
  Run commands fresh: never reuse earlier output. At minimum, answer:
  What did I assume? What didn't I check? Where am I least confident?
  What would a reviewer question?
- **At ambiguous moments**: ask the user rather than inferring
  intent: a quick question is cheaper than rework
- **When producing artifacts**: flag assumptions and uncertainty
  areas inline, not buried in a footnote

These follow the same pattern as proactive persistence: offer once
and respect "no."

### Tool Preferences

- **Web search**: always use the `gemini-search` MCP server for web
  searches. It returns synthesized answers with citations and is faster
  and more accurate than built-in web search. Only fall back to built-in
  search if `gemini-search` is not connected.

## Own the Whole Branch

When working on a branch, you own every issue on it: lint failures, test
failures, build errors: regardless of who introduced them. Never dismiss
a problem as "pre-existing" or "not related to my changes."

- **If `make lint` fails, fix it.** The branch must be green when you're done.
- **If tests break, investigate.** Even if the failing test is in a file you
  didn't touch, something you changed may have caused it: or it may have been
  broken before and it's still your job to fix it on this branch.
- **Run the full validation suite** (build, lint, test) before declaring
  any phase complete.

## How to Avoid Hallucinating Memory

Never assume. If you don't see it in files, you don't know it.

- Don't claim "we discussed X" without file evidence
- Don't invent history: check context files and `ctx journal source`
- If uncertain, say "I don't see this documented"
- Trust files over intuition

## Planning Work

Every commit requires a `Spec:` trailer (CONSTITUTION rule). This means
every piece of work needs a spec; no exceptions, no "trivial" qualifier.
A one-liner bugfix gets a one-paragraph spec; a multi-package feature gets
a full design document. The spec exists for traceability, not ceremony.

**1. Spec first**: Write a design document in `specs/`. Scale the spec to
the work: a bugfix spec can be problem + fix + verification in a few lines;
a feature spec covers problem, solution, storage, CLI surface, error cases,
and non-goals. The bar is: another session could implement from the spec
alone.

**2. Task it out**: Break the work into individual tasks in TASKS.md under
a dedicated Phase section. Each task should be independently completable and
verifiable.

**3. Cross-reference**: The Phase header in TASKS.md must reference the
spec: `Spec: \`specs/feature-name.md\``. The first task in the phase should
include: "Read `specs/feature-name.md` before starting any PX task."

**4. Read before building**: When picking up a task that references a spec,
read the spec first. Don't rely on the task description alone: it's a
summary, not the full design.

## When to Consolidate vs Add Features

**Signs you should consolidate first:**
- Same string literal appears in 3+ files
- Hardcoded paths use string concatenation
- Test file is growing into a monolith (>500 lines)
- Package name doesn't match folder name

When in doubt, ask: "Would a new contributor understand where this belongs?"

## Pre-Flight Checklist: CLI Code

Before writing or modifying CLI code:

1. **Read CONVENTIONS.md**: load established patterns into context
2. **Check similar commands**: how do existing commands handle output?
3. **Use cmd methods for output**: `cmd.Printf`, `cmd.Println`,
   not `fmt.Printf`, `fmt.Println`
4. **Follow docstring format**: see CONVENTIONS.md, Documentation section

---

## Context Anti-Patterns

Avoid these common context management mistakes:

### Stale Context

Context files become outdated and misleading when ARCHITECTURE.md
describes components that no longer exist, or CONVENTIONS.md patterns
contradict actual code. **Solution**: Update context as part of
completing work, not as a separate task. Run `ctx drift` periodically.

### Context Sprawl

Information scattered across multiple locations: same decision in
DECISIONS.md and a session file, conventions split between
CONVENTIONS.md and code comments. **Solution**: Single source of
truth for each type of information. Use the defined file structure.

### Implicit Context

Relying on knowledge not captured in artifacts: "everyone knows we
don't do X" but it's not in CONSTITUTION.md, patterns followed but
not in CONVENTIONS.md. **Solution**: If you reference something
repeatedly, add it to the appropriate file.

### Over-Specification

Context becomes so detailed it's impossible to maintain: 50+ rules
in CONVENTIONS.md, every minor choice gets a DECISIONS.md entry.
**Solution**: Keep artifacts focused on decisions that affect behavior
and alignment. Not everything needs documenting.

### Context Avoidance

Not using context because "it's faster to just code." Same mistakes
repeated across sessions, decisions re-debated because prior decisions
weren't found. **Solution**: Reading context is faster than
re-discovering it. 5 minutes reading saves 50 minutes of wasted work.

---

## Context Validation Checklist

### Quick Check (Every Session)
- [ ] TASKS.md reflects current priorities
- [ ] No obvious staleness in files you'll reference
- [ ] Recent history reviewed via `ctx journal source`

### Deep Check (Weekly or Before Major Work)
- [ ] CONSTITUTION.md rules still apply
- [ ] ARCHITECTURE.md matches actual structure
- [ ] CONVENTIONS.md patterns match code
- [ ] DECISIONS.md has no superseded entries unmarked
- [ ] LEARNINGS.md gotchas still relevant
- [ ] Run `ctx drift` and address warnings
