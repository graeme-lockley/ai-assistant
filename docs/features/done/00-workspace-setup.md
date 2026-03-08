# Feature: Workspace setup

**Status: Done**

## Summary

Provide a workspace template, a default workspace path (`~/.ai-assistant.workspace`), and startup behaviour that ensures a complete workspace layout exists. On server startup, if the workspace root does not exist, populate it from the template. If it exists, verify the layout and create any missing directories or placeholder files.

This is a single feature that covers both the initial template and the runtime scaffold.

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §2–3 (layers, directory structure), §4 (core files).
- [context-loader-spec](../../specs/context-loader-spec.md) §7 (context/ directory structure).

## Scope

- **workspace.template**: A directory shipped with the application containing the full initial workspace:
  - Core files: `AGENT.md`, `IDENTITY.md`, `SOUL.md`, `USER.md`, `MEMORY.md`, `SKILLS.md`, `TOOLS.md`, `TASKS.md`, `WORKSPACE.md` — each with minimal placeholder content (section headings per workspace-design §4).
  - Directories: `logs/`, `memory/`, `skills/`, `tools/`, `context/` with `context/indexes/`, `context/routing/`, `context/cache/`.
  - Template location: embedded in binary or a known path relative to the executable.
- **Default workspace path**: `~/.ai-assistant.workspace` when no override is set. Override via `AI_ASSISTANT_WORKSPACE` env var.
- **Startup behaviour**:
  1. Resolve workspace root (env override or default).
  2. If path does not exist: create and populate from workspace.template (recursive copy).
  3. If path exists: verify required directories exist; create any missing ones. Do not overwrite existing files.
- **Workspace root in config**: Server resolves all workspace paths from this root. File tools respect workspace root for workspace file operations.

## Out of scope

- Context loader; reading workspace into the prompt; session log; memory consolidation; fragment index.
- Migrating an existing workspace to a new template layout.

## Acceptance criteria

- [x] workspace.template exists and contains the full layout matching workspace-design §3 and context-loader-spec §7.
- [x] Default workspace root is `~/.ai-assistant.workspace` when no override is set.
- [x] On startup, if workspace root does not exist, it is created and populated from template.
- [x] On startup, if workspace root exists, missing directories are created; existing files are not overwritten.
- [x] Override via `AI_ASSISTANT_WORKSPACE` works; if override path does not exist, populate from template.
- [x] Server resolves workspace paths from the configured root; file tools respect workspace root.

## Builds toward

**Workspace core in prompt** (01) and **Session log** (02) both need a known workspace root.
