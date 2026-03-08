# Feature: Fragment model and index

**Status: Backlog**

## Summary

Parse workspace markdown files into addressable fragments (by heading/section) and build a fragment index (JSONL) under `context/indexes/`. No retrieval or context assembly yet—this feature only defines the fragment schema, parsing rules, and index build/update so the context loader can later query by fragment.

## Spec reference

- [context-loader-spec](../../specs/context-loader-spec.md) §8 (fragment model), §9 (fragment schema), §10 (file fragmentation rules).

## Scope

- **Fragment schema**: Each record has at least: `id`, `source_file`, `section_path`, `fragment_type`, `content`, `tags`, `estimated_tokens`. Optional: `summary`, `entities`, `importance`, `confidence`, `recency_score`, etc. See context-loader-spec §9.
- **Parsing**: Split workspace files (AGENT, IDENTITY, SOUL, USER, MEMORY, TASKS; optionally SKILL.md and TOOL.md under skills/ and tools/) by heading. One fragment per section or subsection; section_path = list of heading names. Fragment type derived from source file (e.g. `agent`, `identity`, `soul`, `user`, `memory`, `tasks`, `skill`, `tool`).
- **Tags**: Initially from section path or source; optional: allow manual or semi-automatic tags in frontmatter or a sidecar file. Minimal v1 can use section_path and source_file as implicit "tags" for matching.
- **Index**: Write fragments to `context/indexes/fragments.jsonl` (one JSON object per line).
- **Index generation**: Explicit CLI command: `ai-assistant index --workspace=...` scans workspace files, parses headings, and writes the index. Also callable as a Go function so the server can rebuild on startup or on demand. No file watcher in v1 — rebuild is manual or on startup.
- **Token estimate**: Rough estimate per fragment (e.g. chars/4 or a small tokeniser) for later budget trimming.

## Out of scope

- Retrieval, ranking, or composing a working context bundle; classification; entity extraction; routing cards.

## Dependencies

- **Workspace setup** (00): Workspace and `context/indexes/` exist.

## Acceptance criteria

- [ ] All core workspace markdown files (and optionally skills/tools) are parsed into fragments with section_path and content.
- [ ] Fragment records conform to schema (required fields); optional fields supported where applicable.
- [ ] Index is written to `context/indexes/fragments.jsonl`; format is one JSON object per line.
- [ ] Index can be rebuilt via `ai-assistant index` CLI command and programmatically on server startup.
- [ ] estimated_tokens (or equivalent) is present for each fragment for future budget use.

## Builds toward

**Context loader v1** will read the index and select fragments for the working context bundle.
