# Context loader (fragments)

> **Note:** The original multi-phase context-loader specification (retrieval, ranking, routing cards, skip-retrieval) described the Go-era design. The TypeScript codebase implements a **subset**: markdown → fragment records in `context/indexes/fragments.jsonl` for Ring-1/2 source files and `skills/*/SKILL.md`.

## Implemented (`packages/core/src/indexer/indexer.ts`)

- Split selected workspace files by `##` / `###` headings into fragments.
- Each line in `fragments.jsonl` is JSON with at least: `id`, `source_file`, `source_path`, `section_path`, `fragment_type`, `content`, `estimated_tokens`, timestamps.
- Rebuild when the index is missing or any source file is newer than the index.

## Not implemented (legacy spec)

- `context/config.yaml` budgets and ranking weights
- `context/routing/skills.json` / `tools.json` generation
- Skip-retrieval heuristics and semantic search

Those may be reintroduced as follow-up features; this document is retained as a conceptual anchor.
