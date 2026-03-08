# Feature: Skill and tool routing cards

**Status: Backlog**

## Summary

Introduce routing cards for skills and tools: lightweight JSON metadata (name, summary, tags, path, when_to_use / use_when) stored under `context/routing/skills.json` and `context/routing/tools.json`. Skills are described by markdown in `skills/` with a summary in `SKILLS.md`; tools are built into the agent but described in workspace `tools/` and `TOOLS.md`. This feature adds the routing card format and generation or maintenance so the context loader can later select relevant capabilities.

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §5 (capability system), §3 (skills/, tools/, SKILLS.md, TOOLS.md).
- [context-loader-spec](../../specs/context-loader-spec.md) §7 (routing/), §11 (skill routing card), §12 (tool routing card).

## Scope

- **Skill routing cards**: One entry per skill in `skills/`. Each skill lives in `skills/<name>/` with a `SKILL.md` (purpose, actions/elements). `context/routing/skills.json` is an array (or map) of cards: `name`, `summary`, `tags`, `path` (e.g. `skills/architecture-design/SKILL.md`); optional `when_to_use`, `when_not_to_use`, `cost_hint`. Cards can be generated from SKILL.md frontmatter or a separate manifest; or hand-maintained.
- **Tool routing cards**: Tools are implemented in code; workspace holds descriptions. `context/routing/tools.json` lists each tool: `name`, `summary`, `tags`, `path` (e.g. `tools/gog/TOOL.md`); optional `use_when`, `requires_auth`, `risk_level`. TOOL.md files under `tools/<name>/` describe the tool for the loader and agent.
- **SKILLS.md / TOOLS.md**: Top-level summaries (one line or short para per capability) with a link or reference to the full spec. Used for human readability; loader may use routing cards only for selection.
- **Build/update**: Routing cards are updated when skills or tools are added/changed—e.g. on index rebuild, or via a small CLI that scans `skills/` and `tools/` and writes `skills.json` / `tools.json`.

## Out of scope

- Context loader actually using the cards to select capabilities (that is Context loader v2).
- Implementing new tools or skills; only the metadata and file layout for selection.

## Dependencies

- **Workspace scaffold**: `skills/`, `tools/`, `context/routing/` exist.

## Acceptance criteria

- [ ] `context/routing/skills.json` exists and lists each skill under `skills/` with name, summary, tags, path.
- [ ] `context/routing/tools.json` exists and lists each tool with name, summary, tags, path (TOOL.md in workspace).
- [ ] SKILLS.md and TOOLS.md at workspace root provide short summaries of capabilities.
- [ ] Routing cards can be regenerated from workspace (e.g. script or built-in command that scans skills/ and tools/ and optionally SKILL.md/TOOL.md content).

## Builds toward

**Context loader v2** will use these cards to select relevant skills and tools and include them in the working context bundle.
