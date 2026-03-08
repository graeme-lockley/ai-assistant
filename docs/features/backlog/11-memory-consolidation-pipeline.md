# Feature: Memory consolidation pipeline (offline)

**Status: Backlog**

## Summary

Implement the "sleep" consolidation pipeline: a separate job (script or service) that runs offline (e.g. via cron at night), reads session logs and the memory hierarchy, produces hierarchical summaries (daily → weekly → monthly → annual) under `memory/`, and distils high-signal content into MEMORY.md using prompts (e.g. LLM calls). Context loader never reads raw logs or low-level summaries; it only uses MEMORY.md and the fragment index built from it.

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §6 (memory/, flow, sleep consolidation).
- [context-loader-spec](../../specs/context-loader-spec.md) §3 (non-goals: consolidation), §15 (exclusions: logs/, memory/daily/…).

## Scope

- **Inputs**: `logs/` (session transcripts); optionally existing `memory/daily/`, `memory/weekly/`, etc.
- **Pipeline stages**: (1) Summarise logs into daily summaries (e.g. one file per day or per session batch); (2) aggregate daily → weekly, weekly → monthly, monthly → annual; (3) distill from memory hierarchy + recent logs into MEMORY.md (facts, preferences, active threads, beliefs, open questions). Distillation uses prompts (e.g. "Extract durable facts and preferences from these summaries…").
- **Output**: Updated `memory/daily/`, `memory/weekly/`, `memory/monthly/`, `memory/annual/`; updated `MEMORY.md`. MEMORY.md is the only memory artifact the context loader reads at runtime.
- **Execution**: Standalone script or CLI (e.g. `ai-assistant consolidate --workspace=...`) or a separate cron job. Requires workspace path and LLM/config for distillation prompts. No change to the main server or context loader; loader already excludes logs and memory/* from retrieval.
- **Idempotency and safety**: Consolidation should not corrupt existing MEMORY.md on failure; prefer write to temp then rename, or append-only sections with a merge step.

## Out of scope

- Real-time or per-turn consolidation; rewriting USER.md or IDENTITY.md; automatic retention/rotation of logs (can be a separate policy).

## Dependencies

- **Workspace scaffold**, **Session log** (logs/ populated).
- **Fragment model and index** may need to re-index MEMORY.md after consolidation so the loader sees new fragments.

## Acceptance criteria

- [ ] A script or command runs offline and reads from workspace logs/ and memory/.
- [ ] Pipeline produces or updates daily/weekly/monthly/annual summaries under memory/.
- [ ] Pipeline distils into MEMORY.md using configurable prompts (e.g. LLM); MEMORY.md sections align with workspace-design §4 (Facts, Preferences, Active Threads, Beliefs, Open Questions).
- [ ] Context loader does not load from logs/ or memory/daily|weekly|monthly|annual/; only MEMORY.md (and its fragment index) is used at request time.
- [ ] Consolidation is safe (no corrupt state on partial failure) and can be run on a schedule (e.g. cron).

## Builds toward

Complete long-term memory story; agent becomes more relevant over time as MEMORY.md is refreshed from logs and hierarchy.
