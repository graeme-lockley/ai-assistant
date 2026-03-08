# Feature: Agent writeback (TASKS)

**Status: Backlog**

## Summary

Formalise how the agent updates TASKS.md: add newly implied open tasks, close completed tasks, reprioritise. The agent may use existing file tools (write_file, merge_file) or a dedicated "update TASKS" mechanism that the runtime applies after the turn, with optional validation. Ensures workspace-design §4 (TASKS.md) and context-loader-spec §23 (suitable writebacks) are implemented.

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §4 (TASKS.md), §9 (update TASKS when necessary).
- [context-loader-spec](../../specs/context-loader-spec.md) §23 (post-turn writeback, suitable writebacks).

## Scope

- **Allowed updates**: Add new task items; mark tasks complete or remove them; reorder or reprioritise. Format of TASKS.md is implementation-defined (e.g. markdown list, YAML frontmatter, or structured blocks).
- **Mechanism**: (1) Agent uses read_file + write_file or merge_file on workspace TASKS.md (path under workspace root), or (2) agent emits structured "task updates" (e.g. in a tool call or a dedicated response block) that the runtime applies to TASKS.md. Option (1) works with current tool collection; option (2) allows validation and atomic apply.
- **When**: Updates occur after the model responds, during or after post-turn processing. Session log append remains separate.
- **Safety**: Only TASKS.md under the workspace is writable via this mechanism; no direct writes to IDENTITY, SOUL, USER, MEMORY by the agent in normal flow.

## Out of scope

- Automatic extraction of "implied tasks" from the conversation (can be heuristic or LLM in a later feature); full task lifecycle UI.

## Dependencies

- **Workspace scaffold**, **Skip-retrieval and continuation** (post-turn hook).
- **Tool collection** (if using write_file/merge_file for TASKS.md).

## Acceptance criteria

- [ ] Agent can add, complete, or reorder tasks in TASKS.md (via file tools or dedicated update API).
- [ ] Updates are applied to workspace TASKS.md only; no other workspace core files are mutated by this feature.
- [ ] Next turn's context loader can include updated TASKS.md fragments (index may need refresh or on-read parsing).

## Builds toward

Full workspace lifecycle: **Memory consolidation pipeline** handles MEMORY.md; identity evolution remains a separate product decision.
