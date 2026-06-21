# Docs link check validates local PR artifacts

## Problem

The Docs Link Check workflow validates generated documentation in pull requests
against `https://spike.ist`. When a pull request introduces new pages, those
pages exist in the checked-out/generated `docs/` tree but are not published to
the live site until after merge/deploy. Lychee then reports false-positive 404s
for internal links that point at the production origin.

In PR #287, the report contained 1307 errors, mostly repeated 404s for newly
generated recipe pages. The cache marker in the report showed that stale results
were reused, but the root cause was the workflow checking the live site instead
of local artifacts.

## Goals

- Keep pull request link checks focused on the checked-out/generated `docs/`
  tree.
- Preserve production-origin link checking behavior for external links.
- Preserve production-origin link checking behavior for pushes to `main`.
- Keep the workflow output artifact behavior unchanged, and keep sticky PR
  comments for PRs where `GITHUB_TOKEN` has permission to write comments.
- Reference issue #288 for traceability.

## Non-goals

- Do not change the generated documentation content.
- Do not change the production deployment workflow.
- Do not disable lychee failures globally by accepting 404 responses.
- Do not add documentation generation to this workflow; `docs/` is a committed
  artifact and must be updated by documentation PRs before link checking runs.
- Do not use `pull_request_target` to comment on fork PRs from this workflow;
  the link checker runs on PR contents and should stay on `pull_request`.

## Plan

1. Split the link-check behavior by event:
   - pull requests use a generated CI config that remaps internal production
     URLs to the checked-out `docs/` artifact;
   - pushes to `main` keep the existing production-origin behavior with
     `--base-url https://spike.ist`.
2. Generate the pull-request lychee config at runtime using
   `${GITHUB_WORKSPACE}` and pass it explicitly with `--config`; do not rely on
   auto-discovery of `.lychee.toml`.
3. In the pull-request config, use documented lychee config syntax:
   `root_dir = "${GITHUB_WORKSPACE}/docs"`, `index_files = ["index.html"]`, and
   `remap = [...]` entries for:
   - `https://spike.ist` and `https://spike.ist/` -> local `docs/index.html`;
   - `https://spike.ist/...` -> local `docs/...`;
   - root-relative paths like `/main.css` -> local `docs/main.css`;
   - file-style URLs that already point inside a `docs/` tree -> local
     `docs/...`.
4. Keep retry, timeout, accepted status, and report artifact behavior intact.
5. Separate cache keys by event/mode so cached production 404s cannot be reused
   by pull-request local-artifact checks.
6. Include workflow/config/docs source paths in triggers so changes that affect
   link checking run the workflow.
7. Set explicit workflow permissions for repository-origin PR comments, and skip
   the sticky comment step for fork PRs because fork PR `GITHUB_TOKEN`s are
   read-only. Fork PRs still upload the lychee report artifact.
8. Validate by checking workflow syntax and confirming generated remap targets
   cover `https://spike.ist`, `https://spike.ist/`, clean URLs such as
   `https://spike.ist/recipes/backup-and-restore/`, asset URLs, and root-relative
   links such as `/main.css`.
