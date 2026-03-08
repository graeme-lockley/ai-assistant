# context-loader-spec.md
Context Loader Specification for a Workspace-Based AI Agent

---

## 1. Purpose

This document specifies the **Context Loader** for a workspace-based AI agent. The workspace architecture (files, layers, rings) is defined in [workspace-design](workspace-design.md); this spec and that document should be kept in sync.

The Context Loader is responsible for assembling the **smallest sufficient working context** for each request. It prevents prompt bloat, improves relevance, preserves room for reasoning, and helps maintain long-term behavioural coherence. As the workspace and interaction history grow over time, the loader's job is to keep the agent **relevant to the current request**—selecting the right fragments from an ever-larger store so that the agent becomes more useful with time rather than drowning in context.

It sits between:

- the persistent workspace on disk
- the current user request
- the main reasoning model

Its function is analogous to **attention** in cognition: it decides what matters now.

---

## 2. Goals

The Context Loader must:

- load the minimum context required for a turn
- preserve stable agent identity
- retrieve relevant user and memory context
- select only needed skills and tools
- avoid loading raw logs and low-signal summaries
- fit all loaded context into a bounded token budget
- support future evolution toward better ranking and retrieval

The loader should optimise for:

- relevance
- compactness
- stability
- composability
- predictability

---

## 3. Non-Goals

The Context Loader is not responsible for:

- maintaining raw session logs
- performing long-term memory consolidation
- rewriting identity files
- executing tools directly
- replacing the main reasoning model
- loading the entire workspace by default

Memory consolidation belongs to the offline "sleep" pipeline. Tool execution belongs to the runtime orchestration layer.

---

## 4. Position in the Runtime Pipeline

The Context Loader runs before the main model call.

High-level sequence:

1. receive current user request
2. classify request
3. extract entities and themes
4. retrieve candidate fragments
5. rank and filter candidates
6. assemble final working set
7. invoke main model
8. write post-turn task and log updates
9. append session log for later consolidation

Pipeline sketch:

```text
user request
   ↓
context loader
   ├─ classify
   ├─ retrieve
   ├─ rank
   └─ compose
   ↓
main model
   ↓
response
   ↓
task/log updates
```

---

## 5. Context Model

The loader operates using a **three-ring model**.

### Ring 1: Core Self

Always loaded. Order reflects gravity (earlier wins in conflict): SOUL → AGENT → IDENTITY.

These define the stable operating self of the agent:

- `SOUL.md`
- `AGENT.md`
- `IDENTITY.md`

These should be small, high-signal, and curated. See [workspace-design](workspace-design.md) for file roles.

### Ring 2: Supporting Context

Loaded selectively, depending on the request:

- relevant fragments from `USER.md`
- relevant fragments from `MEMORY.md`
- relevant open items from `TASKS.md`

### Ring 3: Capability Context

Loaded only when needed:

- selected skill routing cards
- selected tool routing cards
- full selected `SKILL.md` files
- full selected `TOOL.md` files

Skills and tools must not be loaded wholesale.

---

## 6. Inputs and Outputs

### Inputs

The loader consumes:

- current user message
- optionally recent conversation turns
- workspace files
- fragment index
- routing metadata for skills and tools
- loader configuration
- current token budget

### Output

The loader produces a **Working Context Bundle**.

This bundle contains:

- core files
- selected supporting fragments
- selected capability fragments
- metadata on why each fragment was chosen
- token accounting
- optional execution hints for the orchestrator

Suggested shape:

```json
{
  "core": [],
  "supporting": [],
  "capabilities": [],
  "budget": {
    "max_tokens": 12000,
    "used_tokens": 2860,
    "reserved_for_response": 5000
  },
  "hints": {
    "candidate_skills": ["architecture-design"],
    "candidate_tools": []
  }
}
```

---

## 7. Directory Structure

Recommended additions:

```text
workspace/
  context/
    CONTEXT-LOADER.md
    config.yaml
    indexes/
      fragments.jsonl
      fragments.sqlite
    routing/
      skills.json
      tools.json
    cache/
      recent_queries.json
```

### Purpose of each item

- `CONTEXT-LOADER.md`: human-readable policy and design notes
- `config.yaml`: retrieval, ranking, and budget configuration
- `indexes/fragments.jsonl`: canonical fragment records
- `indexes/fragments.sqlite`: optional searchable store
- `routing/skills.json`: lightweight skill cards
- `routing/tools.json`: lightweight tool cards
- `cache/recent_queries.json`: optional runtime optimisation cache

---

## 8. Fragment Model

The loader should work with **fragments**, not whole files.

A fragment is an addressable unit of content, usually a section or subsection.

Examples:

- `USER.md > Preferences`
- `MEMORY.md > Active Threads`
- `TASKS.md > Open Tasks`
- `skills/architecture-design/SKILL.md > when-to-use`

### Why fragment-based loading matters

Whole-file loading degrades over time because:

- files grow
- unrelated content rides along
- prompt quality drops
- token cost increases

Fragment-based retrieval allows precision.

---

## 9. Fragment Schema

Recommended fragment schema:

```json
{
  "id": "mem-active-threads-001",
  "source_file": "MEMORY.md",
  "source_path": "workspace/MEMORY.md",
  "section_path": ["Active Threads"],
  "fragment_type": "memory",
  "authority": "derived",
  "content": "The user is designing an OpenClaw-inspired agent workspace.",
  "summary": "Current work on agent workspace design.",
  "tags": ["agent", "workspace", "architecture", "memory"],
  "entities": ["OpenClaw", "workspace"],
  "importance": 0.84,
  "confidence": 0.93,
  "recency_score": 0.79,
  "stability_score": 0.82,
  "evidence_count": 4,
  "last_reinforced_at": "2026-03-08",
  "created_at": "2026-03-08",
  "updated_at": "2026-03-08",
  "estimated_tokens": 42
}
```

### Required fields

- `id`
- `source_file`
- `section_path`
- `fragment_type`
- `content`
- `tags`
- `estimated_tokens`

### Recommended fields

- `importance`
- `confidence`
- `recency_score`
- `stability_score`
- `evidence_count`
- `summary`
- `entities`

---

## 10. File Fragmentation Rules

### AGENT.md
Fragment by major heading:
- mission
- principles
- constraints
- operating rules

### IDENTITY.md
Fragment by:
- identity
- purpose
- domain
- strengths
- limitations

### SOUL.md
Fragment by:
- disposition
- tone
- values
- style

### USER.md
Fragment by:
- identity
- preferences
- interests
- projects
- constraints
- recurring relationships

### MEMORY.md
Fragment by:
- facts
- preferences
- active threads
- reinforced beliefs
- cautions
- open questions

### TASKS.md
Fragment by task item or small groups of related tasks.

### SKILL.md and TOOL.md
Each should expose:
- routing card
- full detailed spec

---

## 11. Skill Routing Card Specification

The loader should search over **routing cards** first, not full skill bodies.

Recommended schema:

```json
{
  "name": "architecture-design",
  "summary": "Create structured architecture documents and file-by-file design specifications.",
  "tags": ["architecture", "design", "specification", "systems"],
  "when_to_use": [
    "User asks for a system design",
    "User asks for a specification",
    "User wants a structured technical document"
  ],
  "when_not_to_use": [
    "User is making casual small talk"
  ],
  "cost_hint": "medium",
  "path": "skills/architecture-design/SKILL.md"
}
```

### Required fields

- `name`
- `summary`
- `tags`
- `path`

---

## 12. Tool Routing Card Specification

The loader should treat tools similarly.

Recommended schema:

```json
{
  "name": "gog",
  "summary": "Access Gmail for reading and sending email.",
  "tags": ["gmail", "email", "communication"],
  "use_when": [
    "User asks to read email",
    "User asks to send or draft email"
  ],
  "requires_auth": true,
  "risk_level": "medium",
  "path": "tools/gog/TOOL.md"
}
```

### Required fields

- `name`
- `summary`
- `tags`
- `path`

---

## 13. Classification Phase

The first stage of the loader is request classification.

The goal is not perfect taxonomy. It is to determine what kinds of context are likely to matter.

Possible request classes:

- conversation
- reflection
- personal writing
- planning
- research
- architecture design
- coding
- debugging
- documentation
- tool-use request
- memory retrieval
- task follow-up

A request may have multiple classes.

Example:

> "Please design a memory system for my agent and write a spec"

Classification:

- architecture design
- documentation
- memory subsystem
- likely skill retrieval
- likely user/memory retrieval

### Implementation options

Good initial approaches:

- rules + keywords
- simple heuristics
- tiny classifier model
- lightweight LLM pass

Start simple.

---

## 14. Entity and Theme Extraction

After classification, extract entities and topics from the request.

Examples:

- project names
- domains
- artifact types
- people
- technologies
- explicit file names
- explicit verbs like design, compare, summarise, draft

Example input:

> "Please design the bytecode format for my VM"

Possible extraction:

- entities: VM, bytecode
- themes: compiler, language design, architecture
- artifact: specification

These extractions guide retrieval.

---

## 15. Retrieval Sources

Candidate fragments may come from:

- `SOUL.md` (always)
- `AGENT.md` (always)
- `IDENTITY.md` (always)
- `USER.md`
- `MEMORY.md`
- `TASKS.md`
- skill routing cards
- tool routing cards
- full skill/tool specs after selection

### Important exclusions

Never retrieve directly from:

- `logs/`
- `memory/daily/`
- `memory/weekly/`
- `memory/monthly/`
- `memory/annual/`

These are source material for consolidation, not runtime prompt context.

---

## 16. Retrieval Strategy

A practical starter strategy:

### Step 1: Always load core self
Always include (in order): `SOUL.md`, `AGENT.md`, `IDENTITY.md`.

### Step 2: Search candidate fragments
Search structured fragments in:

- `USER.md`
- `MEMORY.md`
- `TASKS.md`
- skill routing cards
- tool routing cards

### Step 3: Build candidate pool
Candidate pool may be generated using:

- keyword matching
- tag overlap
- semantic similarity
- entity matches
- recency weighting

### Step 4: Rank and trim
Select the best non-duplicative fragments within the budget.

---

## 17. Ranking Model

Each fragment should receive a composite score.

Suggested scoring function:

```text
score =
  (relevance_weight * semantic_relevance)
+ (importance_weight * importance)
+ (recency_weight * recency_score)
+ (confidence_weight * confidence)
+ (stability_weight * stability_score)
+ (entity_match_weight * entity_overlap)
- (cost_weight * token_cost_penalty)
- (duplication_weight * duplication_penalty)
```

### Typical ranking factors

- semantic similarity to request
- keyword overlap
- tag overlap
- explicit entity match
- fragment importance
- confidence
- recency
- stability
- token cost
- duplication with already-selected fragments

### Hard suppression conditions

Suppress fragments that are:

- stale and low-importance
- contradicted by newer memory
- near-duplicates of already-selected fragments
- too expensive relative to marginal value

---

## 18. Token Budget Policy

The loader must enforce a hard context budget.

### Example working budget

For a 16k input budget:

- 1.5k core self
- 2.0k supporting context
- 1.0k routing cards
- 1.5k full skill/tool details if needed
- 10.0k reserved for conversation, reasoning, and output

The exact numbers depend on model context and orchestration policy.

### Core principle

The loader must preserve enough room for the model to reason and respond. A perfect retrieval set that leaves no room for thought is a bad retrieval set.

---

## 19. Selection Rules

Suggested default rules:

### Always include
- `SOUL.md`
- `AGENT.md`
- `IDENTITY.md`

### Usually include
- 1 to 3 `USER.md` fragments
- 1 to 3 `MEMORY.md` fragments

### Conditionally include
- up to 2 `TASKS.md` fragments
- up to 3 skill routing cards
- up to 2 tool routing cards

### Load full skill/tool specs only when selected
For example:
- when a skill meaningfully improves execution
- when a tool is likely to be used
- when the tool or skill has safety-relevant usage rules

---

## 20. Working Context Bundle Format

Recommended bundle format:

```yaml
core:
  - source: SOUL.md
    reason: always_load
  - source: AGENT.md
    reason: always_load
  - source: IDENTITY.md
    reason: always_load

supporting:
  - source: USER.md
    section: Preferences
    reason: relevant_to_requested_output_style
  - source: MEMORY.md
    section: Active Threads
    reason: matches_agent_workspace_topic

capabilities:
  - source: routing/skills.json
    item: architecture-design
    reason: task_is_system_design

full_specs:
  - source: skills/architecture-design/SKILL.md
    reason: selected_skill

budget:
  total: 16000
  used: 2940
  reserved_for_reasoning: 10000
```

This structure also improves observability and debugging.

---

## 21. Skip-Retrieval Cases

Not every turn needs full retrieval.

Examples where the loader may load only the core self:

- greetings
- acknowledgements
- short social replies
- isolated rewrite requests with user-provided content
- trivial clarifications unrelated to prior context

This lowers latency and cost.

---

## 22. Continuation Cases

Some turns clearly continue prior work. In those cases, the loader should include relevant `TASKS.md` fragments (and possibly more MEMORY or USER context).

Examples:

- "continue"
- "expand that"
- "rewrite section 4"
- "now produce the companion document"
- "apply those changes"

Continuation detection is important for multi-step artifact workflows.

---

## 23. Post-Turn Writeback

The loader itself usually does not write durable memory. After the main model responds, the agent may update:

- `TASKS.md`
- session log

### Suitable writebacks

`TASKS.md` (updated by the agent when necessary):
- add newly implied open tasks
- close completed tasks
- reprioritise active tasks

### Unsuitable writebacks
Do not write directly to:
- `IDENTITY.md`
- `SOUL.md`
- `USER.md` without evidence/curation
- `MEMORY.md` outside the consolidation pipeline, unless your design explicitly supports incremental memory with validation

---

## 24. Conflict Handling

Memory and user context may conflict.

Examples:
- older preference vs newer preference
- multiple active projects
- stale assumptions from past sessions

Recommended precedence:

1. explicit current user message
2. authoritative current files
3. high-confidence recent memory
4. older derived memory

Conflict handling rules:

- prefer fresh evidence
- prefer explicit user correction
- suppress contradicted fragments
- mark unresolved contradictions for consolidation

---

## 25. Suggested Loader Configuration

Example `context/config.yaml`:

```yaml
always_load:
  - AGENT.md
  - IDENTITY.md
  - SOUL.md

never_load_paths:
  - logs/**
  - memory/daily/**
  - memory/weekly/**
  - memory/monthly/**
  - memory/annual/**

max_fragments:
  user: 3
  memory: 3
  tasks: 2
  skill_cards: 3
  tool_cards: 2

ranking_weights:
  semantic_relevance: 0.35
  importance: 0.20
  recency: 0.10
  confidence: 0.10
  stability: 0.05
  entity_overlap: 0.15
  token_cost_penalty: 0.03
  duplication_penalty: 0.02

budgets:
  input_total_tokens: 16000
  core_max_tokens: 1500
  supporting_max_tokens: 2500
  routing_max_tokens: 1000
  full_specs_max_tokens: 1500
  reserved_for_reasoning_and_output: 9500
```

---

## 26. Go-Friendly Data Structures

Suggested Go types:

```go
type Fragment struct {
    ID               string   `json:"id"`
    SourceFile       string   `json:"source_file"`
    SourcePath       string   `json:"source_path"`
    SectionPath      []string `json:"section_path"`
    FragmentType     string   `json:"fragment_type"`
    Authority        string   `json:"authority"`
    Content          string   `json:"content"`
    Summary          string   `json:"summary"`
    Tags             []string `json:"tags"`
    Entities         []string `json:"entities"`
    Importance       float64  `json:"importance"`
    Confidence       float64  `json:"confidence"`
    RecencyScore     float64  `json:"recency_score"`
    StabilityScore   float64  `json:"stability_score"`
    EvidenceCount    int      `json:"evidence_count"`
    EstimatedTokens  int      `json:"estimated_tokens"`
    LastReinforcedAt string   `json:"last_reinforced_at"`
    CreatedAt        string   `json:"created_at"`
    UpdatedAt        string   `json:"updated_at"`
}

type WorkingContextBundle struct {
    Core         []Fragment       `json:"core"`
    Supporting   []Fragment       `json:"supporting"`
    Capabilities []Fragment       `json:"capabilities"`
    FullSpecs    []Fragment       `json:"full_specs"`
    Budget       BudgetReport     `json:"budget"`
    Hints        ExecutionHints   `json:"hints"`
}

type BudgetReport struct {
    TotalTokens            int `json:"total_tokens"`
    UsedTokens             int `json:"used_tokens"`
    ReservedForReasoning   int `json:"reserved_for_reasoning"`
}

type ExecutionHints struct {
    CandidateSkills []string `json:"candidate_skills"`
    CandidateTools  []string `json:"candidate_tools"`
}
```

---

## 27. Go-Friendly Pseudocode

```text
function BuildWorkingContext(request, recentConversation, workspace, config):

    core = loadAlwaysFiles(workspace, config.always_load)

    classification = classifyRequest(request, recentConversation)

    themes = extractThemes(request)
    entities = extractEntities(request)

    candidates = []

    candidates += retrieveFragments(USER.md, themes, entities)
    candidates += retrieveFragments(MEMORY.md, themes, entities)
    candidates += retrieveFragments(TASKS.md, themes, entities)

    if isContinuation(request, recentConversation):
        # Emphasise TASKS and possibly more MEMORY/USER context
        candidates += retrieveFragments(TASKS.md, themes, entities)

    skillCards = retrieveSkillCards(themes, entities, classification)
    toolCards  = retrieveToolCards(themes, entities, classification)

    candidates += skillCards
    candidates += toolCards

    ranked = rankCandidates(candidates, request, config.ranking_weights)

    selected = fitToBudget(
        core,
        ranked,
        config.budgets,
    )

    selectedFullSpecs = []

    for each selected capability in selected.capabilities:
        if shouldLoadFullSpec(capability, classification):
            selectedFullSpecs += loadFullSpec(capability)

    bundle = composeBundle(core, selected.supporting, selected.capabilities, selectedFullSpecs)

    return bundle
```

---

## 28. Minimal Viable Implementation

A practical version 1 should use:

- heading-based file fragmentation
- JSONL fragment index
- keyword + tag matching
- simple heuristic ranking
- fixed token budgets
- routing cards for skills and tools

This is enough to create a useful system.

### Version 1 priorities

1. structure files into headings
2. parse them into fragments
3. attach tags manually or semi-automatically
4. implement simple retrieval
5. apply budget trimming
6. emit a working context bundle

---

## 29. Later Enhancements

After the minimal version is working, consider:

- embeddings for semantic retrieval
- SQLite or Tantivy/Bleve-backed search
- reinforcement scoring from repeated mentions
- contradiction detection
- decayed recency weighting
- domain-specific retrieval policies
- traceability and introspection UI
- offline evaluation of retrieval quality
- automatic routing-card generation from skill/tool specs

---

## 30. Observability and Debugging

The loader should produce diagnostics.

Useful diagnostics include:

- selected fragments and reasons
- rejected fragments and reasons
- token usage by source
- final score breakdown
- loaded skills/tools
- continuation detection result

These logs are invaluable when the agent seems distracted, repetitive, or forgetful.

Example debug entry:

```json
{
  "fragment_id": "mem-active-threads-001",
  "selected": true,
  "score": 0.87,
  "reason": "High semantic match to request and reinforced recently.",
  "estimated_tokens": 42
}
```

---

## 31. Safety and Stability Considerations

The loader contributes to behavioural stability.

Benefits:

- reduces accidental identity drift
- reduces irrelevant old memory intrusion
- prevents whole-workspace prompt stuffing
- preserves room for deliberate reasoning

Hard rules:

- do not auto-load raw logs
- do not auto-load low-level summary archives
- do not mutate identity files as part of normal turn handling
- do not prefer recency over explicit user correction

---

## 32. Recommended Default Policy

A sensible default policy is:

- always load `AGENT.md`, `IDENTITY.md`, `SOUL.md`
- retrieve up to 2 relevant user fragments
- retrieve up to 2 relevant memory fragments
- include tasks only for continuation or explicit workflow
- search skill/tool routing cards broadly
- load full specs narrowly
- preserve at least 50% of budget for reasoning and output

This will give a compact, stable, scalable baseline.

---

## 33. Summary

The Context Loader is the agent's **attention mechanism over its own workspace**.

It decides what the agent should remember now, rather than forcing the whole self into every prompt. As the workspace grows—more memory, richer user model, additional skills—the loader keeps each turn focused on the minimum context needed to be relevant. That allows the agent to **become more relevant over time**: more history and capability do not degrade behaviour if retrieval and ranking are sound.

A good loader:

- protects coherence
- reduces token waste
- improves relevance
- scales with workspace growth
- makes a minimal system prompt viable
- supports continued development (new skills, growing USER.md and MEMORY.md) without prompt bloat

In a workspace-based agent architecture, the Context Loader is not an optional optimisation. It is a core subsystem. See [workspace-design](workspace-design.md) for the workspace architecture and how it evolves over time.
