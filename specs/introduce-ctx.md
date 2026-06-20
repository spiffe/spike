# Spec: Introduce optional ctx project context to SPIKE

## Problem Statement

SPIKE is introducing `ctx` as an optional, project-scoped AI context
layer. The initial `ctx init` plus a GitNexus indexing run produced a
useful starting point, but left the repository's agent-facing files in
an inconsistent state:

- The GitNexus code-intelligence block was duplicated: embedded inside
  `CLAUDE.md` and also the sole content of `AGENTS.md`.
- `AGENTS.md` carried tool-specific instructions instead of pointing to
  the canonical agent guide.
- `ctx`- and GitNexus-generated artifacts (`Makefile.ctx`,
  `GETTING_STARTED.md`, `.claude/`, `.context/ingest/`) were not aligned
  with project conventions or ignore rules.
- A stale `CLAUDE.md.<ts>.bak` backup was left behind.

The goal is a clean, convention-aligned layout that mirrors the
reference `ctx` repository, while making clear that the AI tooling is
optional for contributors.

## Proposed Solution

Mirror the layout used by the reference `ctx` repository:

1. **Single source of agent truth.** `CLAUDE.md` remains the canonical
   guide. `AGENTS.md` reduces to a pointer: "Read and follow CLAUDE.md."
2. **Extract GitNexus content.** Move the GitNexus block out of
   `CLAUDE.md` into a standalone `GITNEXUS.md`, retaining the
   `<!-- gitnexus:start/end -->` markers so `gitnexus analyze` can
   regenerate it in place. `CLAUDE.md` keeps a short "Companion Tools"
   pointer to it.
3. **Optional framing.** Both `CLAUDE.md` and `GITNEXUS.md` state that
   GitNexus is optional and only relevant "if installed." `GITNEXUS.md`
   carries a project-owned preamble *above* the managed markers so it
   survives regeneration.
4. **Align the Makefile.** Move `Makefile.ctx` to `makefiles/Ctx.mk`
   (matching the `makefiles/*.mk` PascalCase convention) and update the
   root `Makefile` to `-include ./makefiles/Ctx.mk`.
5. **Ignore generated/local artifacts.** Add `GETTING_STARTED.md`,
   `Makefile.ctx` (stray regenerated copy), `.context/ingest/`, the
   whole `.claude/` directory, and the local scratch notebooks
   `ideas/`, `inbox/`, and `outbox/` to `.gitignore`.
6. **Contributor docs.** Add an "Optional: AI-Assisted Development
   Tooling" section to `CONTRIBUTING.md` covering `ctx` and GitNexus
   install steps, explicitly marked optional.
7. **Remove cruft.** Delete `CLAUDE.md.<ts>.bak`.

## File Surface

- `CLAUDE.md` (modified): GitNexus block replaced by Companion Tools
  pointer.
- `AGENTS.md` (modified): reduced to a CLAUDE.md pointer.
- `GITNEXUS.md` (new): extracted GitNexus guide + owned preamble.
- `makefiles/Ctx.mk` (moved from `Makefile.ctx`).
- `Makefile` (modified): include path updated.
- `CONTRIBUTING.md` (modified): optional tooling section.
- `.gitignore` (modified): new ignore entries.
- `.context/**` (new): committed project context used by ctx, excluding
  generated and local material. Tracked: the canonical files
  (`CONSTITUTION.md`, `CONVENTIONS.md`, `DECISIONS.md`, `LEARNINGS.md`,
  `TASKS.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, the playbooks),
  `templates/`, and `handovers/.gitkeep`. Untracked via `.gitignore`:
  `journal*`, `logs`, `state/`, `ingest/`, `.ctx.key`, and
  `handovers/*` (timestamped notes). Tracked files must hold durable,
  project-relevant context only; both humans and agents may edit them.
  Local notes, transcripts, scratch state, machine-specific paths, and
  (per CONSTITUTION) any credentials or customer data must never be
  committed.

## Error / Edge Cases

- **ctx regenerates `Makefile.ctx` at root.** Neutralized: the root
  `-include` is removed, so a stray regenerated copy injects no targets,
  and the path is gitignored. Cost: `makefiles/Ctx.mk` no longer
  auto-tracks ctx upstream changes; it is now project-owned. See
  DECISIONS.
- **GitNexus regenerates `GITNEXUS.md`.** Only the marked block is
  regenerated; the owned preamble above the start marker is preserved.
- **`.claude/` ignored entirely.** Nothing under it is tracked today
  (verified via `git ls-files .claude`), so no history is lost.
- **A future `ctx init` rewrites agent-facing files.** The project-owned
  layout remains authoritative. Regenerated root-level helper files are
  either ignored or manually reconciled into the project-owned
  locations.

## Non-Goals

- No changes to SPIKE application code, build, or test behavior.
- Not adopting the KB editorial workflow (`.context/kb/` absent);
  `.context/ingest/` scaffolding is ignored until that workflow is used.

## Verification

- `git check-ignore` confirms the new ignore entries match.
- `make` parses with the updated include (no duplicate-target errors).
- `CLAUDE.md` and `AGENTS.md` contain no `gitnexus:` markers.
- `git status --ignored` shows generated/local ctx and Claude artifacts
  are ignored as expected, while intended `.context/` files remain
  tracked.
