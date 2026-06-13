# Knowledge Base

The kb is the project's evidence-tracked knowledge surface.
It lives at `.context/kb/` and is governed by the editorial
contract in `../ingest/KB-RULES.md`. Topic pages live under
`topics/<slug>/index.md`. Cross-cutting artifacts
(`glossary.md`, `domain-decisions.md`, `contradictions.md`,
`outstanding-questions.md`, `evidence-index.md`,
`source-coverage.md`, `timeline.md`) live alongside this
file.

---

## Scope

<!--
TODO: declare what this kb covers, in one paragraph.

`ctx kb ingest`, `ctx kb ask`, `ctx kb ground`, and
`ctx kb site-review` all refuse to run while this placeholder
remains in place. Replace the entire HTML comment with a
prose paragraph that:

  - names the domain the kb covers;
  - names what the kb deliberately does NOT cover;
  - names the operator audience (the human and any agent
    skills that read this).

Keep it to one paragraph. Specificity matters more than
length.
-->

---

## Topics

<!-- CTX:KB:TOPICS START -->
<!--
This block is managed by `ctx kb reindex`. Do not hand-edit
between the START and END markers. Content here is
regenerated from `.context/kb/topics/*/index.md` and any
manual edits are overwritten on the next reindex.

Each topic page registered by `ctx kb topic new` appears here
as a bullet linking to `topics/<slug>/index.md` with the
topic's lede sentence.

Until a topic exists, this block stays empty.
-->
<!-- CTX:KB:TOPICS END -->

---

## Conventions

This kb is governed by:

- **`../ingest/KB-RULES.md`** is the editorial contract:
  pass-mode discipline, topic-page circuit breaker,
  source-coverage state machine, topic-adjacency pre-flight,
  cold-reader rubric, life-stage check, evidence discipline,
  confidence bands, demotion policy, closeout shape.
- **`../ingest/schemas/`** holds field-level shape for each
  cross-cutting artifact (`evidence-index.md`, `glossary.md`,
  `contradictions.md`, `outstanding-questions.md`,
  `domain-decisions.md`, `timeline.md`, `source-map.md`,
  `source-coverage.md`, `relationship-map.md`). Each schema
  ships shape, not content.
- **`../../specs/kb-editorial-pipeline.md`** is the full spec,
  including the failure-analysis section and the v1
  non-goals. Read this when you need to understand *why* a
  rule exists, not just *what* it says.

The mode skills (`/ctx-kb-ingest`, `/ctx-kb-ask`,
`/ctx-kb-site-review`, `/ctx-kb-ground`, `/ctx-kb-note`) are
the canonical writers. Hand-edits to kb files are an escape
hatch, not the primary path.
