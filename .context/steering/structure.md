---
name: structure
description: Project structure and directory conventions
inclusion: always
priority: 10
---
<!--
  This is a ctx steering file: persistent behavioral
  rules prepended to prompts based on the frontmatter
  above.

  inclusion (when the rule fires):
    always  → injected on EVERY tool call. On Claude
              Code this is the only mode that fires
              AUTOMATICALLY (the PreToolUse hook
              passes an empty prompt to ctx agent).
              Use for invariants and for any rule
              that MUST fire reliably.
    auto    → injected when the prompt matches the
              `description` field above.
                - Cursor / Cline / Kiro: native.
                - Claude Code: Claude calls the
                  ctx_steering_get MCP tool on its
                  own when it decides the rule is
                  relevant. The ctx plugin ships
                  the MCP auto-registration; verify
                  with `claude mcp list`.
    manual  → only when the file is explicitly named
              (e.g. via the MCP tool or a skill).

  priority (ordering within a tier):
    Lower numbers inject first. 10 is a reasonable
    default for invariants; use 50 for normal rules.

  tools (scope the rule to specific AI tools):
    Empty list = applies to all tools (default).
    Example:  tools: [claude, cursor]

  Edit the body below, then delete this comment.
  See docs/cli/steering.md for the full reference.
-->

# Project Structure

<!-- remove this after you edit the steering file !-->

Describe the project layout and directory conventions.

- **Top-level directories and their purpose**
- **Where new files should go** (and where they should not)
- **Naming conventions** for files, packages, modules