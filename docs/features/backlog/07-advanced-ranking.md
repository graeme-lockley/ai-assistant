# Feature: Advanced ranking and observability

**Status: Backlog**

## Summary

Add a configurable ranking model, per-source limits, ranking weights, and observability so the loader selects the best fragments within budget using composite scoring. Basic token budgets (hard caps, 50% reserve) are already in Context loader v1 (04); this feature adds the sophistication: configurable weights, max_fragments per source, ranked selection with reasons, and introspectable selection logs.

## Spec reference

- [context-loader-spec](../../specs/context-loader-spec.md) §17 (ranking), §18 (token budget — advanced config), §19 (selection rules), §25 (config), §30 (observability), §32 (default policy).

## Scope

- **Config**: `context/config.yaml` with: `always_load` (file list), `never_load_paths` (e.g. logs, memory/daily/…), `max_fragments` per source (user, memory, tasks, skill_cards, tool_cards), `ranking_weights` (e.g. semantic_relevance, importance, recency, token_cost_penalty). Feature 04 already reads basic budget values from config; this feature adds the ranking and per-source tuning knobs.
- **Composite scoring**: Each candidate fragment gets a composite score based on configurable weights: keyword/tag overlap, recency, importance, minus token cost penalty. Sort by score; select in order until budget is full. Hard suppression: drop stale/low-importance, near-duplicates, or fragments that exceed marginal value.
- **Budget trimming (advanced)**: After ranking, fit selected fragments into per-category and total budgets; trim from the bottom of the ranked list. Per-source max_fragments prevents any single source from dominating.
- **Observability**: Loader emits metadata for each selected fragment (id, score, breakdown, reason) and optionally rejected fragments with their scores. Token usage by source. Useful for debugging "agent seems distracted or forgetful." Output format: structured log or returned alongside the bundle.

## Out of scope

- Embeddings or semantic similarity (later enhancement — see 11); SQLite/Tantivy search; automatic routing-card generation; observability UI (see 11).

## Dependencies

- **Context loader v2** (06): Full pipeline with classification and capability selection.
- **Context loader v1** (04): Basic budgets already enforced; this feature enhances ranking within those budgets.

## Acceptance criteria

- [ ] `context/config.yaml` supports `max_fragments` per source and `ranking_weights`; loader reads and applies them.
- [ ] Candidates are ranked by a composite score using configurable weights; selection respects per-source limits and total budgets.
- [ ] Loader emits or logs selected fragments with score breakdown and reason; optionally rejected fragments and scores.
- [ ] Ranking can be tuned without code changes (config only).

## Builds toward

**Skip-retrieval and continuation** (08) and **Agent writeback (TASKS)** (09) complete the loader behaviour; **Context loader enhancements** (11) can add semantic retrieval and better ranking later.
