# Feature: Token budget and ranking

**Status: Backlog**

## Summary

Introduce configurable token budgets (core, supporting, routing, full_specs, reserved for reasoning) and a simple ranking model so the loader selects the best fragments within budget. Add loader config (e.g. `context/config.yaml`) and observability (selected/rejected fragments and reasons).

## Spec reference

- [context-loader-spec](../../specs/context-loader-spec.md) §17 (ranking), §18 (token budget), §19 (selection rules), §25 (config), §30 (observability), §32 (default policy).

## Scope

- **Config**: `context/config.yaml` (or equivalent) with: `always_load` (file list), `never_load_paths` (e.g. logs, memory/daily/…), `max_fragments` per source (user, memory, tasks, skill_cards, tool_cards), `ranking_weights` (e.g. semantic_relevance, importance, recency, token_cost_penalty), `budgets` (input_total_tokens, core_max, supporting_max, routing_max, full_specs_max, reserved_for_reasoning_and_output).
- **Ranking**: Each candidate fragment gets a composite score (e.g. keyword/tag overlap, recency, importance, minus token cost). Sort by score; select in order until budget is full. Hard suppression: drop stale/low-importance, near-duplicates, or fragments that exceed marginal value.
- **Budget trimming**: After ranking, fit selected fragments into per-category and total budgets; trim from the bottom of the list or drop lowest-scoring until within limits. Always preserve reserved_for_reasoning (e.g. 50% of input budget).
- **Observability**: Log or return metadata for each selected fragment (id, score, reason) and optionally rejected fragments; token usage by source. Useful for debugging "agent seems distracted or forgetful."

## Out of scope

- Embeddings or semantic similarity (later enhancement); SQLite/Tantivy search; automatic routing-card generation.

## Dependencies

- **Context loader v2**: Full pipeline with classification and capability selection.

## Acceptance criteria

- [ ] Loader reads config from `context/config.yaml` (or env/defaults); budgets and max_fragments are configurable.
- [ ] Candidates are ranked by a composite score; selection respects per-category and total token budgets.
- [ ] At least 50% of input budget (or configured reserve) is preserved for conversation and response.
- [ ] Loader emits or logs selected fragments with reason and token usage; optional: rejected fragments and scores.

## Builds toward

**Skip-retrieval and continuation** and **Agent writeback (TASKS)** complete the loader behaviour; **Context loader enhancements** can add semantic retrieval and better ranking later.
