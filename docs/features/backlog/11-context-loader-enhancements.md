# Feature: Context loader enhancements (optional later)

**Status: Backlog**

## Summary

Optional improvements to the context loader after the minimal viable implementation is in place: semantic retrieval (embeddings), better search (SQLite/Tantivy/Bleve), reinforcement scoring, contradiction detection, decayed recency, domain-specific policies, observability UI, and automatic routing-card generation. Implement as needed; not required for "full" workspace + loader implementation.

## Spec reference

- [context-loader-spec](../../specs/context-loader-spec.md) §29 (later enhancements), §30 (observability).

## Scope (pick per sub-feature)

- **Semantic retrieval**: Embed fragments and the request; retrieve by vector similarity instead of or in addition to keyword/tag. Requires embedding model and index (e.g. vector store).
- **Structured search**: Replace or augment JSONL fragment index with SQLite or Tantivy/Bleve for faster and richer queries (e.g. full-text, filters).
- **Reinforcement and decay**: Score fragments by "reinforced" mentions over time; apply recency decay so very old, low-importance fragments rank lower.
- **Contradiction detection**: Detect when memory or user context conflicts; prefer fresh evidence and suppress contradicted fragments (see context-loader-spec §24).
- **Observability UI**: Introspection view of selected/rejected fragments, token usage, and reasons per turn (traceability).
- **Automatic routing cards**: Generate or update skills.json/tools.json from SKILL.md/TOOL.md content (e.g. parse "when to use" sections).

## Out of scope for MVP

- Required for "full implementation" of workspace-design and context-loader-spec: the previous features (scaffold through memory consolidation) are sufficient. This feature is a placeholder for incremental improvements.

## Dependencies

- **Context loader v2**, **Token budget and ranking** (and optionally **Skip-retrieval and continuation**).

## Acceptance criteria (per chosen sub-feature)

- [ ] Each enhancement is optional and configurable; core loader behaviour remains valid when enhancements are disabled.
- [ ] Semantic retrieval (if implemented): fragments and requests are embedded; retrieval uses similarity; token and latency impact documented.
- [ ] Observability (if implemented): selected/rejected fragments and token breakdown available for debugging or UI.

## Builds toward

Ongoing improvement of relevance and scalability as the workspace grows; no further "required" features for the spec.
