# Learnings

<!--
UPDATE WHEN:
- Discover a gotcha, bug, or unexpected behavior
- Debugging reveals non-obvious root cause
- External dependency has quirks worth documenting
- "I wish I knew this earlier" moments
- Production incidents reveal gaps

DO NOT UPDATE FOR:
- Well-documented behavior (link to docs instead)
- Temporary workarounds (use TASKS.md for follow-up)
- Opinions without evidence
-->

<!-- INDEX:START -->
| Date | Learning |
|----|--------|
| 2026-06-13 | Zola 0.19+/0.22 moved syntax highlighting config and renamed themes |
<!-- INDEX:END -->

<!-- Add gotchas, tips, and lessons learned here -->
## [2026-06-13-123540] Zola 0.19+/0.22 moved syntax highlighting config and renamed themes

**Context**: make docs failed on Zola 0.22.1: 'unknown field highlight_code' then 'Theme base16-ocean-dark does not exist'. docs-src/config.toml used pre-0.19 highlighting keys.

**Lesson**: Zola 0.19 moved highlight_code/highlight_theme out of [markdown] into a [markdown.highlighting] table (theme/light_theme/dark_theme/style). Zola 0.22 swapped Syntect for the Giallo highlighter, so Syntect theme names like base16-ocean-dark are gone; valid names come from the Giallo bundle (getzola/giallo, sourced from shikijs/textmate-grammars-themes), e.g. material-theme-ocean, one-dark-pro, github-dark, nord.

**Application**: When a Zola upgrade breaks 'make docs', migrate config.toml highlighting into [markdown.highlighting] and map old theme names to Giallo identifiers. List valid themes via: gh api repos/getzola/giallo/readme --jq .content | base64 -d. base16-ocean-dark -> material-theme-ocean is the closest match for SPIKE.
