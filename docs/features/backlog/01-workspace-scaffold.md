# Feature: Workspace scaffold

**Status: Backlog**

## Summary

Introduce the workspace directory structure and core file placeholders so the agent has a persistent, on-disk "cognitive substrate." No context loading or prompt assembly yet—this feature only creates the layout and ensures the server can resolve a workspace root. Default workspace path and population from **workspace.template** on first run are defined in [00-workspace-template](00-workspace-template.md).

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §2–3 (layers, directory structure), §4 (core files).

## Scope

- **Workspace root**: Configurable path (e.g. env `AI_ASSISTANT_WORKSPACE` or `AI_ASSISTANT_ROOT_DIR`); default is `~/.ai-assistant.workspace` (see 00-workspace-template). Server and any future loader resolve all workspace paths from this root.
- **Directory layout**: Create or recognise:
  - Top-level: `AGENT.md`, `IDENTITY.md`, `SOUL.md`, `USER.md`, `MEMORY.md`, `SKILLS.md`, `TOOLS.md`, `TASKS.md`, `WORKSPACE.md`.
  - Directories: `logs/`, `memory/`, `skills/`, `tools/`, `context/` (and under `context/`: `indexes/`, `routing/`, `cache/` as per context-loader-spec).
- **Placeholders**: If a core file or directory is missing, the system may create empty or minimal placeholders (e.g. empty files or single-line stubs) so the layout is complete. Optional: bootstrap only on first use.
- **No loader**: This feature does not read workspace content into the prompt. The agent may continue to use a fixed system prompt or existing behaviour.

## Dependencies

- **00-workspace-template**: Provides workspace.template and default path `~/.ai-assistant.workspace` with copy-from-template on first run. Scaffold uses that default when no override is set and may assume the layout was initialised from the template.

## Out of scope

- Context loader; fragment index; session log appends; memory consolidation.
- Populating core files with real content (user/operator responsibility or later features).

## Acceptance criteria

- [ ] Workspace root is configurable (env or config); server resolves workspace paths from it.
- [ ] Required directories and core files exist (created on demand or at startup if missing).
- [ ] File tools (read_file, read_dir, write_file) respect workspace root for workspace paths (or a dedicated workspace path) so the agent can read/write workspace files when that is added.
- [ ] Layout matches workspace-design §3 and context-loader-spec §7 for `context/`.

## Builds toward

Enables **Workspace core in prompt** and **Session log** (both need a known workspace root and `logs/`).
