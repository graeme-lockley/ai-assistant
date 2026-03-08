# Feature: Skip-retrieval and continuation

**Status: Backlog**

## Summary

Optimise loader behaviour for light-weight turns (skip full retrieval) and for continuation turns (emphasise TASKS and recent context). After the model responds, ensure post-turn writeback (TASKS.md updates and session log append) is triggered. Implements context-loader-spec §21–23.

## Spec reference

- [context-loader-spec](../../specs/context-loader-spec.md) §21 (skip-retrieval), §22 (continuation), §23 (post-turn writeback).

## Scope

- **Skip-retrieval**: When the request is classified as a greeting, acknowledgement, short social reply, or trivial clarification unrelated to prior context, load only Ring 1 (core self). No supporting or capability retrieval. Lowers latency and cost.
- **Continuation detection**: When the request clearly continues prior work (e.g. "continue", "expand that", "rewrite section 4", "apply those changes"), set a continuation flag. In continuation mode: include relevant TASKS.md fragments and optionally more MEMORY/USER context; ensure TASKS are ranked higher or given more budget.
- **Post-turn writeback**: After the main model responds: (1) Append the turn to the session log (already in Session log feature). (2) Allow the agent to update TASKS.md—e.g. via a tool (write_file/merge_file on TASKS.md) or structured output that the runtime applies. Loader itself does not write; runtime or agent applies updates. Document the contract: "agent may update TASKS.md when necessary."

## Out of scope

- Loader mutating IDENTITY, SOUL, USER, or MEMORY (only consolidation pipeline or explicit user flow may do that).
- Full implementation of "agent infers new tasks" (can be heuristic or LLM-suggested in a later iteration).

## Dependencies

- **Context loader v2** and **Token budget and ranking** (so retrieval and budgets exist).
- **Session log** (append already implemented).

## Acceptance criteria

- [ ] For skip-retrieval request types, loader returns only core (Ring 1); no supporting or capability context.
- [ ] For continuation requests, loader emphasises TASKS and optionally more MEMORY/USER; continuation is detectable (e.g. keyword or simple classifier).
- [ ] Post-turn: session log is appended; TASKS.md can be updated by the agent (via existing file tools or a dedicated TASKS update path).
- [ ] Loader/runtime never writes to IDENTITY, SOUL, USER, or MEMORY as part of normal turn handling.

## Builds toward

**Agent writeback (TASKS)** can formalise how the agent proposes TASKS.md changes; **Memory consolidation pipeline** handles MEMORY.md offline.
