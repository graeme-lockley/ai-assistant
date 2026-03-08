# Backlog

Planned features for the AI Assistant, in rough priority order. Workspace and context-loader features are ordered so each increment builds on the previous and leads to full implementation (see [workspace-design](../specs/workspace-design.md) and [context-loader-spec](../specs/context-loader-spec.md)).

## Backlog

| # | Feature | Spec | Status |
|---|---------|------|--------|
| 0 | **Workspace template** — workspace.template as initial workspace; server defaults to ~/.ai-assistant.workspace and populates from template if missing. | [00-workspace-template.md](backlog/00-workspace-template.md) | Backlog |
| 1 | **Workspace scaffold** — Directory structure, core file placeholders, configurable workspace root. | [01-workspace-scaffold.md](backlog/01-workspace-scaffold.md) | Backlog |
| 2 | **Workspace core in prompt** — Load SOUL, AGENT, IDENTITY (and optionally USER, MEMORY, TASKS) as whole files into system prompt; priority order and size cap. | [02-workspace-core-in-prompt.md](backlog/02-workspace-core-in-prompt.md) | Backlog |
| 3 | **Session log** — Append each turn to workspace logs/; append-only; never loaded into prompt by default. | [03-session-log.md](backlog/03-session-log.md) | Backlog |
| 4 | **Fragment model and index** — Parse workspace markdown into fragments by heading; JSONL index under context/indexes/. | [04-fragment-model-and-index.md](backlog/04-fragment-model-and-index.md) | Backlog |
| 5 | **Context loader v1** — Core (Ring 1) always; simple retrieval for USER, MEMORY, TASKS from index; emit working context bundle; build system prompt from bundle. | [05-context-loader-v1.md](backlog/05-context-loader-v1.md) | Backlog |
| 6 | **Skill and tool routing cards** — context/routing/skills.json and tools.json; SKILLS.md, TOOLS.md; card format and build from skills/ and tools/. | [06-skill-and-tool-routing-cards.md](backlog/06-skill-and-tool-routing-cards.md) | Backlog |
| 7 | **Context loader v2** — Classification and entity/theme extraction; select skills and tools via routing cards; include capabilities and full specs in bundle. | [07-context-loader-v2.md](backlog/07-context-loader-v2.md) | Backlog |
| 8 | **Token budget and ranking** — config.yaml budgets and max_fragments; rank candidates; fit to budget; observability (selected/rejected, token usage). | [08-token-budget-and-ranking.md](backlog/08-token-budget-and-ranking.md) | Backlog |
| 9 | **Skip-retrieval and continuation** — Skip full retrieval for greetings/ack; continuation detection and TASKS emphasis; post-turn session log + TASKS writeback contract. | [09-skip-retrieval-and-continuation.md](backlog/09-skip-retrieval-and-continuation.md) | Backlog |
| 10 | **Agent writeback (TASKS)** — Agent can add, complete, reorder tasks in TASKS.md via file tools or dedicated update; only TASKS.md mutated. | [10-agent-writeback-tasks.md](backlog/10-agent-writeback-tasks.md) | Backlog |
| 11 | **Memory consolidation pipeline** — Offline job (cron): logs → daily/weekly/monthly/annual summaries → distill into MEMORY.md with prompts. | [11-memory-consolidation-pipeline.md](backlog/11-memory-consolidation-pipeline.md) | Backlog |
| 12 | **Context loader enhancements** — Optional: semantic retrieval, SQLite/Tantivy, reinforcement/contradiction, observability UI, auto routing cards. | [12-context-loader-enhancements.md](backlog/12-context-loader-enhancements.md) | Backlog |

## Done

| Feature | Spec |
|---------|------|
| **Control plane** — Slash commands from REPL: /exit (close session, then quit), /models (list models), /model (set/query session model), /help (client-side list of commands). Models hardcoded for v1. | [control-plane.md](done/control-plane.md) |
| **Tool collection** — Web search, web get, exec bash, read/write/merge file, read dir. Tools respect workspace constraints. | [tool-collection.md](done/tool-collection.md) |
| **REPL history** — Readline-style history in the REPL; Up/Down to navigate history, Left/Right within the line. History persisted across sessions. | [repl-history.md](done/repl-history.md) |
| **Streaming results** — All results streamed to the caller (no single-chunk responses). Request/response support multiple content types; session ID; HTTP with SSE/NDJSON; LLM client disables gzip and streams reasoning_content; REPL flushes stdout per token. | [streaming-results.md](done/streaming-results.md) |
| **Session console output** — Log to server console (with timestamp) when a session is created and when it is closed. Sessions have a defined lifecycle (created → active → closed). Explicit close via `X-Session-Close: true`. | [session-console-output.md](done/session-console-output.md) |

---

*Add new rows above; move items to Done or Backlog as needed.*
