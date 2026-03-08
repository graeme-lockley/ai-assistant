# Feature: Session log

**Status: Backlog**

## Summary

Append each turn (user message and assistant response) to a session log under the workspace `logs/` directory. Logs are append-only, never loaded into the prompt in normal operation, and used only for later summarisation and offline consolidation. Implements workspace-design §6 (logs/).

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §6 (logs/).
- [context-loader-spec](../../specs/context-loader-spec.md) §4 (pipeline: append session log), §23 (post-turn writeback).

## Scope

- **Where**: Under workspace `logs/`. One file per session (e.g. `logs/<session-id>.log` or `logs/YYYY-MM-DD-<session-id>.md`). Format: append-only; each turn appends the user message and the full assistant response (and optionally timestamps, turn IDs).
- **When**: After the main model responds and the server has the complete response; append before returning to the client (or in a fire-and-forget goroutine with minimal error handling).
- **Content**: At minimum: user message and assistant response. Optional: session ID, turn index, timestamp. No raw logs are loaded into the context loader or prompt unless explicitly requested (e.g. future "include last N turns" option).
- **Config**: Workspace root must be set (from Workspace scaffold). If workspace or `logs/` is not configured, session logging can be a no-op.

## Out of scope

- Reading logs into the prompt; consolidation or summarisation; retention or rotation policy (can be added later).

## Dependencies

- **Workspace scaffold**: Workspace root and `logs/` directory.

## Acceptance criteria

- [ ] Each completed turn is appended to a session log file under `workspace/logs/`.
- [ ] Log file is keyed by session (e.g. session ID); format is append-only and human-readable.
- [ ] Session log is not read by the context loader or system prompt builder in normal operation.
- [ ] If workspace is not configured, session log is skipped without error.

## Builds toward

**Memory consolidation pipeline** will consume logs to produce summaries and feed MEMORY.md.
