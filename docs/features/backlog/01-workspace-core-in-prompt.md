# Feature: Workspace core in prompt (whole-file)

**Status: Backlog** (Ring 1 implemented: SOUL, AGENT, IDENTITY loaded into system prompt on each session create.)

**Note: Stepping stone.** Replaced by Context loader v1 (04) once the fragment model exists. Purpose: get workspace identity into the prompt immediately before fragmentation is built.

## Summary

Load the workspace "core self" (and optionally supporting files) into the system prompt as whole files, in the priority order defined in workspace-design. No fragmentation yet—files are loaded in full up to a size cap or truncated. Delivers the minimal system prompt from workspace-design §9 so the agent has identity, role, and (optionally) user/memory context.

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §8 (context rings), §9 (minimal system prompt).
- [context-loader-spec](../../specs/context-loader-spec.md) §5 (Ring 1: Core Self).

## Scope

- **Ring 1 (always)**: Read `SOUL.md`, `AGENT.md`, `IDENTITY.md` from workspace root. Concatenate in that order into the system prompt. If a file is missing, skip or use empty string.
- **Ring 2 (optional for v1)**: Optionally append `USER.md` and `MEMORY.md` (and/or `TASKS.md`) in order, each up to a fixed character or token cap (e.g. first N chars or 500 tokens). Can be disabled by config for minimal v1.
- **Cap / truncation**: Enforce a total system-prompt size limit (e.g. 2k–4k tokens); truncate from the end of the concatenation if over. Or per-file caps so core stays intact.
- **Order and gravity**: Prompt text must state the priority order (SOUL → AGENT → IDENTITY → USER → MEMORY) and that earlier wins in conflict.
- **Integration**: The component that builds the request to the LLM (e.g. agent or a small "prompt builder") takes workspace root and config, reads the files, and returns the system prompt string. No fragment index or retrieval yet.

## Out of scope

- Fragment model; retrieval; ranking; skill/tool routing; token budget policy; skip-retrieval or continuation.

## Dependencies

- **Workspace setup** (00): Workspace root and core files must exist.

## Acceptance criteria

- [x] System prompt includes SOUL.md, AGENT.md, IDENTITY.md in that order when present. (Ring 1 done via workspace.LoadBootstrap in session create.)
- [ ] Optional: USER.md and MEMORY.md (and TASKS.md) included with a size cap.
- [ ] Total system prompt size is bounded (configurable); overflow truncates from the end.
- [ ] Prompt text reflects workspace-design §9 (priority order and rules).

## Builds toward

**Fragment model and index** and **Context loader v1** will replace whole-file load with fragment-based retrieval; this feature gives an immediate behaviour change (identity/role in prompt) before fragmentation exists.
