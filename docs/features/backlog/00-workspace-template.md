# Feature: Workspace template and default workspace

**Status: Backlog**

## Summary

Provide a **workspace.template** that defines the initial workspace layout and placeholder content. On server startup, the workspace root defaults to **`~/.ai-assistant.workspace`**. If that directory does not exist, the server populates it by copying from workspace.template so the user gets a ready-to-use workspace without manual setup.

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §2–3 (layers, directory structure), §4 (core files).

## Scope

- **workspace.template**: A directory (or archive) shipped with the application that contains the full initial workspace:
  - Core files: `AGENT.md`, `IDENTITY.md`, `SOUL.md`, `USER.md`, `MEMORY.md`, `SKILLS.md`, `TOOLS.md`, `TASKS.md`, `WORKSPACE.md` (each with minimal placeholder content or section headings as per workspace-design §4).
  - Directories: `logs/`, `memory/`, `skills/`, `tools/`, `context/` with `context/indexes/`, `context/routing/`, `context/cache/` (empty or with stub files as needed).
  - Template location: embedded in the binary, or under a known path relative to the executable (e.g. `workspace.template/` in the repo, copied into the build or deployed alongside the server).
- **Default workspace path**: The server uses **`~/.ai-assistant.workspace`** as the default workspace root when no override is set (e.g. env `AI_ASSISTANT_WORKSPACE` or `AI_ASSISTANT_ROOT_DIR` unset). `~` is expanded per OS (e.g. `$HOME` on Unix).
- **Startup behaviour**: On server startup:
  1. Resolve the workspace root (config/env or default `~/.ai-assistant.workspace`).
  2. If the workspace root path does not exist: create the directory and populate it by copying the contents of workspace.template into it (recursive copy of files and subdirectories). Do not overwrite if the directory already exists (first-run only).
  3. If the workspace root already exists: use it as-is; do not replace or overwrite user content.
- **Config override**: An explicit workspace path (env or config) overrides the default; if that path does not exist, the same “populate from template” behaviour may apply, or the server may require the directory to exist (implementation choice; recommend: populate from template when path is missing and is the default, otherwise fail or document that user must create it).

## Out of scope

- Context loader; reading workspace into the prompt; session log; memory consolidation. This feature is only template + default path + copy-on-first-run.
- Migrating or upgrading an existing workspace to a new template layout (future feature).

## Acceptance criteria

- [ ] workspace.template exists and contains the full layout (core files + directories) matching workspace-design §3 and context-loader-spec §7.
- [ ] Default workspace root is `~/.ai-assistant.workspace` when no override is set.
- [ ] On server startup, if the workspace root does not exist, the server creates it and populates it by copying from workspace.template.
- [ ] If the workspace root already exists, the server uses it without modifying it.
- [ ] Override via config/env (e.g. `AI_ASSISTANT_WORKSPACE`) still works; behaviour when override path does not exist is defined (e.g. populate from template or exit with clear error).

## Builds toward

**Workspace scaffold** (01) builds on this: it assumes a workspace root (now defaulting to `~/.ai-assistant.workspace` and optionally created from template) and ensures the layout is present and the server resolves paths correctly.
