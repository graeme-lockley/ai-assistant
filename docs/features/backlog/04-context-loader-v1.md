# Feature: Context loader v1 (core + simple retrieval)

**Status: Backlog**

## Summary

Implement a first version of the context loader: always load Ring 1 (SOUL, AGENT, IDENTITY); for Ring 2, select a small number of fragments from USER, MEMORY, and TASKS using the fragment index and simple matching (keyword or tag overlap). Emit a working context bundle with token accounting and enforce a hard token budget. Replaces whole-file loading from feature 01.

## Spec reference

- [workspace-design](../../specs/workspace-design.md) §7–8 (context loader, rings).
- [context-loader-spec](../../specs/context-loader-spec.md) §5 (rings), §6 (bundle), §16 (retrieval strategy), §20 (bundle format), §28 (minimal viable implementation).

## Scope

- **Ring 1**: Always load full content of SOUL.md, AGENT.md, IDENTITY.md (or their fragment equivalents from the index if already fragmented). Order: SOUL → AGENT → IDENTITY.
- **Ring 2**: Query fragment index for USER, MEMORY, TASKS. Simple retrieval: keyword match on request text and/or tag overlap with request (e.g. extract a few words from the user message; match fragment tags or section_path). Select up to N fragments per source (e.g. 2–3 USER, 2–3 MEMORY, 2 TASKS) with a total cap (e.g. 2k tokens for supporting).
- **Bundle**: Produce a WorkingContextBundle with `core`, `supporting`, and `budget` (used/reserved). No `capabilities` or full_specs in v1.
- **Token budget (basic)**: Enforce a hard total token budget. Reserve at least 50% for conversation + response. Per-category caps: core (1.5k), supporting (2.5k). If over budget, drop lowest-scoring supporting fragments. Read budget config from `context/config.yaml` or defaults. A loader without budgets is broken — this is not optional.
- **Compose prompt**: From bundle, concatenate core then supporting fragments into the system prompt within the budget.
- **Integration**: Before calling the LLM, the agent (or server) calls the loader with (request, workspace path, config); loader returns bundle; caller builds system prompt from bundle and invokes model.

## Out of scope

- Request classification; entity/theme extraction; skill/tool routing; ranking model; skip-retrieval; continuation detection; observability UI.

## Dependencies

- **Workspace scaffold**, **Workspace core in prompt** (or equivalent: loader replaces simple whole-file prompt).
- **Fragment model and index**: Index must exist and be populated.

## Acceptance criteria

- [ ] Loader loads SOUL, AGENT, IDENTITY in order every time.
- [ ] Loader retrieves up to N fragments from USER, MEMORY, TASKS using simple keyword/tag match against the current request.
- [ ] Output is a WorkingContextBundle (core + supporting + token accounting).
- [ ] System prompt is built from the bundle and stays within a configured token budget.
- [ ] Agent uses the loader output for each turn (when workspace is configured).

## Builds toward

**Skill and tool routing cards** and **Context loader v2** (classification, capability selection, ranking).
