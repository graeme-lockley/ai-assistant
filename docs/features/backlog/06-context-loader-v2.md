# Feature: Context loader v2 (classification and capability selection)

**Status: Backlog**

## Summary

Add request classification and entity/theme extraction; use routing cards to select relevant skills and tools and include them in the working context bundle. Load full SKILL.md/TOOL.md only when a capability is selected. Composes the full three-ring model (core + supporting + capability context) and integrates with the loader pipeline.

## Spec reference

- [context-loader-spec](../../specs/context-loader-spec.md) §13 (classification), §14 (entity/theme extraction), §5 (Ring 3), §19 (selection rules), §22 (continuation), §27 (pseudocode).

## Scope

- **Classification**: Classify the current user message into request classes (e.g. conversation, coding, documentation, tool-use, task follow-up). Implementation: rules + keywords or simple heuristics. Output used to decide whether to load more supporting context and which capability cards to search.
- **Entity and theme extraction**: Extract entities (project names, technologies, file names) and themes (e.g. "architecture", "debugging") from the request. Use for matching against fragment tags and routing card tags.
- **Ring 3**: Search routing cards (skills.json, tools.json) by tag/entity/theme overlap; select up to N skill cards and M tool cards (e.g. 3 skills, 2 tools). Add selected cards to the bundle. For each selected capability, optionally load the full SKILL.md or TOOL.md and add to bundle (full_specs) when classification suggests it will improve execution (e.g. tool-use request, or "design" → load architecture skill spec).
- **Bundle**: Extend WorkingContextBundle with `capabilities` (routing card content) and `full_specs` (full markdown when loaded). Hints: `candidate_skills`, `candidate_tools` for the orchestrator.
- **Pipeline**: Loader flow is: classify → extract entities/themes → load core → retrieve supporting fragments → retrieve routing cards → rank/trim → optionally load full specs → compose bundle.

## Out of scope

- Sophisticated ranking (that is Token budget and ranking); skip-retrieval and continuation (separate feature); observability UI.

## Dependencies

- **Context loader v1**: Core + supporting retrieval and bundle.
- **Fragment model and index**, **Skill and tool routing cards**.

## Acceptance criteria

- [ ] Request is classified into one or more classes; entities and themes are extracted.
- [ ] Loader selects skill and tool routing cards based on classification and tag/entity match.
- [ ] Selected cards are included in the bundle; full SKILL.md/TOOL.md are loaded when appropriate (e.g. when tool likely to be used or skill improves execution).
- [ ] Working context bundle includes core, supporting, and capabilities (and full_specs when loaded).
- [ ] System prompt is built from the full bundle (order: core → supporting → capabilities/full_specs).

## Builds toward

**Token budget and ranking** will add configurable budgets and ranking so selection stays within token limits and prefers relevant fragments.
