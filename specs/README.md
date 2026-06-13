# Specs

Formalized plans for features, refactors, and non-trivial changes.

A spec is what comes out of a planning session: problem statement,
proposed solution, CLI surface, storage, error cases, and non-goals.
It's complete enough that another session could implement from it alone.

## Lifecycle

1. **Draft**: write the spec in this directory
2. **Reference**: add a Phase to TASKS.md with `Spec: specs/<name>.md`
3. **Implement**: follow the spec, checking off tasks as you go
4. **Archive**: move to `specs/done/` when all tasks are complete

## Tips

- Keep specs concise. A page is usually enough.
- Non-goals are as important as goals: they prevent scope creep.
- If a spec grows beyond two pages, split it.
