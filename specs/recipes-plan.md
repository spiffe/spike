# Spec: SPIKE documentation recipes

## Problem Statement

SPIKE's docs are strong on **reference** (the 73-row env-var table, per-command
pages) and **architecture** (system overview, security model, 32 ADRs), and have
a few **deployment tutorials** (quickstart, bare-metal) and **operations**
how-tos (recovery, backup, production). What they lack is a **task-oriented
"recipes" layer**: "I want to do X — here is the problem, the gist, the
workflow, the gotchas."

This gap is not cosmetic. The single hardest question in the PR that produced
this spec — *what is `lite` mode, and does it need keepers and a root key?* —
is documented only as `memory, lite, sqlite` in one table cell. There is no
page that explains the three backend modes, when to pick each, or that lite is
an encryption-only service that still requires keepers. That missing recipe is
why understanding the system took so long.

## Goal

Add a `docs-src/content/recipes/` section modeled on the ctx recipe format
(<https://ctx.ist/recipes/>): each recipe is a self-contained, task-first page.

### Recipe template (every recipe page)

1. **Problem** — the concrete situation/question, in the reader's words.
2. **TL;DR** — the one-paragraph / few-command answer.
3. **Workflow** — numbered, copy-pasteable steps.
4. **Tips** — non-obvious advice, defaults, recommended values.
5. **Pitfalls** — the traps (e.g., paths are namespaces not filesystem paths;
   patterns are regex not globs; lite still needs keepers).
6. **Cross-links** — related recipes + reference/architecture pages.
7. **What's next** — the logical follow-on recipe.

## Proposed recipe inventory

Ranked; Tier 1 are the gaps that directly caused this PR's confusion.

### Tier 1 — Concepts & decisions (highest value)
1. **Choosing a backend store: memory vs lite vs sqlite.** Decision matrix:
   persistence, root key, keepers required, use case. (memory = dev, volatile,
   no keepers/root key; lite = encryption-only, no persistence, *needs* keepers
   + root key, for external storage e.g. S3/minio; sqlite = production,
   persistent, needs keepers + root key.)
2. **Bootstrapping a fresh SPIKE.** Generate the root key, split into Shamir
   shares, seed the keepers; verification; idempotency (the
   `spike-bootstrap-state` ConfigMap); what to do if it stalls. (Bare-metal
   `make bootstrap` and the Kubernetes bootstrap Job.)
3. **Where the root key lives: keepers, Shamir, and auto-recovery.** Why Nexus
   contacts keepers on startup (`InitializeBackingStoreFromKeepers`), shares vs
   threshold, `SendShardsPeriodically`, and which mode (memory) is the only
   keeper-free one.

### Tier 2 — Day-to-day usage
4. **Storing and reading secrets** (put/get/list/delete/undelete, versions,
   metadata).
5. **Writing access policies** (SPIFFE ID + path regex, permissions,
   `create` vs `apply`, YAML). Pitfall-heavy: regex not globs, namespaced
   paths.
6. **Granting a workload access to secrets** — the end-to-end connective recipe:
   register the SPIRE entry → write the SPIKE policy → workload reads via SDK.
7. **Using SPIKE as an encryption service (cipher + lite mode)** — encrypt/
   decrypt without persisting secrets; the S3/minio pattern.

### Tier 3 — Operations & lifecycle
8. **Break-the-glass disaster recovery** (`operator recover`/`restore`,
   recovery shards) + how to rehearse it safely.
9. **Backup and restore** (sqlite DB + root-key shards).
10. **Deploying on Kubernetes** (helm chart, the spike-* components, the
    bootstrap Job) and the bare-metal counterpart.
11. **Production hardening** (recipe-ize the existing production guide).
12. **Troubleshooting** — symptom-first: "Nexus never becomes Ready / stuck in
    keeper recovery" (exactly this PR's failure), "policy created but access
    denied", "bootstrap won't complete".

### Tier 4 — Integration & advanced
13. **Integrating the Go SDK** (read secrets from an app).
14. **Upgrading SPIKE.**

## Non-Goals

- Rewriting the reference (config table, command pages) or architecture/ADRs;
  recipes link to them.
- Net-new product features.

## Verification

- `make docs` builds clean with the new `recipes/` section.
- Each recipe follows the 7-part template and cross-links to reference pages.

## Checklist (no recipe gets untouched)

File location: `docs-src/content/recipes/<slug>.md`.

- [x] `recipes/_index.md` — section landing
- [x] T1 #1 `choosing-a-backend-store.md`
- [x] T1 #2 `bootstrapping-spike.md`
- [x] T1 #3 `root-key-keepers-recovery.md`
- [x] T2 #4 `storing-and-reading-secrets.md`
- [x] T2 #5 `writing-access-policies.md`
- [x] T2 #6 `granting-a-workload-access.md`
- [x] T2 #7 `encryption-as-a-service.md`
- [x] T3 #8 `break-the-glass-recovery.md`
- [x] T3 #9 `backup-and-restore.md`
- [x] T3 #10 `deploying-spike.md`
- [x] T3 #11 `production-hardening.md`
- [x] T3 #12 `troubleshooting.md`
- [x] T4 #13 `go-sdk-integration.md`
- [x] T4 #14 `upgrading-spike.md`

Notes: section added at `weight = 9` to avoid renumbering other sections and
keep `make docs` green; nav/weight placement can be promoted in a follow-up.
A `toc_recipes()` macro / nav-template entry may be needed for the recipes
section to appear in the site navigation (verify against templates/).
