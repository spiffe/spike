# Tasks

<!--
UPDATE WHEN:
- New work is identified → add task with #added timestamp
- Starting work → add #in-progress or #started timestamp
- Work completes → mark [x]
- Work is blocked → add to Blocked section with reason
- Scope changes → update task description inline

DO NOT UPDATE FOR:
- Reorganizing or moving tasks (violates CONSTITUTION)
- Removing completed tasks (use ctx task archive instead)

STRUCTURE RULES (see CONSTITUTION.md):
- Tasks stay in their Phase section permanently: never move them
- Use inline labels: #in-progress, #blocked, #priority:high
- Mark completed: [x], skipped: [-] (with reason)
- Never delete tasks, never remove Phase headers

TASK STATUS LABELS:
  `[ ]`: pending
  `[x]`: completed
  `[-]`: skipped (with reason)
  `#in-progress`: currently being worked on (add inline, don't move task)
-->

### Phase 1: [Name] `#priority:high`
- [ ] Task 1
- [ ] Task 2

### Phase 2: [Name] `#priority:medium`
- [ ] Task 1
- [ ] Task 2

## Blocked

### Maintenance

- [x] Clear remaining uncalled govulncheck advisories (10 import + 4 module) via a dependency-bump pass #priority:low #session:6d40ae08 #branch:chore/introduce-ctx #commit:ca9b541 #added:2026-06-13-130622 #done:2026-06-13 (x/crypto v0.52.0, grpc v1.79.3; govulncheck now 0 total)
- [ ] Wire `SPIKE_NEXUS_PBKDF2_ITERATION_COUNT` into the crypto path: the constant `NexusPBKDF2IterationCount` is defined (spike-sdk-go config/env/env.go:45) and documented (docs-src/content/usage/configuration.md, default 600000) but no code reads it (no `…Val()` accessor, no os.Getenv). Add the accessor + consumer so the documented option takes effect, or escalate upstream to the SDK. #priority:low #session:00ce042d #branch:chore/introduce-ctx #added:2026-06-15 (found during docs-vs-code config audit)
- [x] Fix Docs Link Check so PRs validate internal `https://spike.ist/...` links against the checked-out `docs/` artifact while pushes to `main` keep production-origin behavior. Spec: `specs/docs-link-check-local-artifacts.md`. Issue: #288. #priority:high #branch:fix/docs-link-check-local-artifacts #added:2026-06-21 #done:2026-06-21
