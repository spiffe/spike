# Spec: Fix docs build for Zola 0.22 (Giallo highlighter)

## Problem Statement

`make docs` fails on Zola 0.22.x:

```
ERROR TOML parse error at line 15, column 1
unknown field `highlight_code`, expected one of `highlighting`, ...
```

`docs-src/config.toml` uses the pre-0.19 syntax-highlighting keys
(`highlight_code`, `highlight_theme`) under `[markdown]`. Zola 0.19
moved highlighting into the `[markdown.highlighting]` table, and Zola
0.22 replaced the Syntect highlighter with Giallo, which uses VSCode /
textmate theme identifiers. The old theme name `base16-ocean-dark` no
longer exists. The published site (`docs/`) cannot be regenerated, so
documentation is effectively unshippable on a current Zola.

## Proposed Solution

1. Migrate `docs-src/config.toml` to the Giallo configuration:
   - Remove `highlight_code` / `highlight_theme` from `[markdown]`.
   - Add a `[markdown.highlighting]` table with
     `theme = "material-theme-ocean"` (the bundled Giallo theme closest
     to the retired `base16-ocean-dark`).
2. Regenerate and publish the static site via `make docs`
   (`build-docs.sh` runs `zola build` and copies `public/` into the
   tracked `docs/` tree; `cover.sh` refreshes the coverage report).

## File Surface

- `docs-src/config.toml` (modified): highlighting migrated.
- `docs/**` (regenerated): ~51 HTML pages re-emitted with the new
  inline highlight colors, plus the refreshed coverage report.

## Error / Edge Cases

- **Theme identifiers changed with Giallo.** Old Syntect names
  (`base16-*`) are invalid; valid names come from the Giallo bundled set
  (`getzola/giallo`, sourced from shikijs/textmate-grammars-themes).
- **Inline vs class styling.** Default `style = "inline"` bakes colors
  into spans, so the theme change produces a large but expected `docs/`
  diff. Not switching to `style = "class"` (would require linking
  generated `giallo*.css` and adding `.giallo-l`/`.giallo-ln` CSS).

## Non-Goals

- No light/dark dual-theme support (single dark theme retained).
- No Zola pinning or toolchain version management in this change.

## Verification

- `zola --root docs-src build` succeeds (59 pages, 0 orphans).
- `make docs` exits 0 with no `ERROR` lines.
