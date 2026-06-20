# Agent Playbook (Gate)

Distilled directives injected at session start. Full playbook:
read AGENT_PLAYBOOK.md when you need behavioral guidance, session
lifecycle details, or anti-patterns.

## Invoke ctx from PATH

```bash
ctx status        # correct
./dist/ctx        # wrong: never hardcode paths
go run ./cmd/ctx  # wrong: unless developing ctx itself
```

## When `ctx` Errors

If the error names your flag, argument, or command, read
`ctx <cmd> --help` and fix the call. Otherwise, relay verbatim
and stop. When unsure, stop.

## File Interaction Protocol

When a task involves reading, modifying, or reasoning about a file:

1. **Read before act**: Do not rely on memory, summaries, or prior reads
2. **No partial reads**: Do not sample and assume the rest
3. **Freshness requirement**: Do not reuse stale context from earlier in the 
   session
4. **Edit authority comes from visibility**: If you haven't seen it, you don't 
   get to modify it
5. **Coverage requirement**: Before editing, state what parts of the file were 
   read and why they are sufficient

## Planning Work

Do not begin implementation without a spec.

Every commit requires a `Spec:` trailer. Every piece of work needs
a spec; no exceptions. Scale the spec to the work.

The design-to-implementation chain is:

```text
/ctx-brainstorm  →  /ctx-plan  →  /ctx-spec  →  /ctx-implement
   (vague)      (contested)    (committed)     (execution)
```

`/ctx-brainstorm` shapes a vague idea into a bet. `/ctx-plan`
attacks the bet and writes a debated brief to
`.context/briefs/<TS>-<slug>.md`. `/ctx-spec` (optionally
`--brief <path>`) absorbs the brief into a committed spec under
`specs/`. Skip the predecessors only when the step's input is
already settled.

## Proactive Persistence

After completing a task, making a decision, or hitting a gotcha,
persist before continuing. Don't wait for session end.

## Chunk and Checkpoint

For multi-step work: commit after each chunk, persist learnings,
run tests before moving on. Track progress via TASKS.md checkboxes.

## Independent Review

A review must occur:

* Before the first code change
* After completing tasks
* Before presenting results

Review must consider:

* Spec
* TASKS.md
* Current implementation

## Tool Preferences

Use the `gemini-search` MCP server for web searches. Fall back to
built-in search only if `gemini-search` is not connected.

## Conversational Triggers

| User Says                                       | Action               |
|-------------------------------------------------|----------------------|
| "Do you remember?" / "What were we working on?" | `/ctx-remember`      |
| "How's our context looking?"                    | `/ctx-status`        |
| "What should we work on?"                       | `/ctx-next`          |
| "Commit this" / "Ship it"                       | `/ctx-commit`        |
| "What did we learn?"                            | `/ctx-reflect`       |
| "Save that as a decision"                       | `/ctx-decision-add`  |
| "That's worth remembering"                      | `/ctx-learning-add`  |
| "Add a task for that"                           | `/ctx-task-add`      |
| "Let's wrap up"                                 | Reflect then persist |
