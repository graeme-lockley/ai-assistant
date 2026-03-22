# Workspace design

## Rings (prompt assembly)

1. **Ring 1 — always**: `SOUL.md`, `IDENTITY.md` (shared across all agents).
2. **Agent body**: `agents/<name>/AGENT.md` markdown body (plus optional skills catalog section injected for allowed skills).
3. **Ring 2 — optional**: `USER.md`, `MEMORY.md`, `TASKS.md` with per-file and global token caps (`AI_ASSISTANT_RING2_MAX_TOKENS`, `AI_ASSISTANT_SYSTEM_PROMPT_MAX_TOKENS`).

## Directory layout

```
<workspace>/
  SOUL.md, IDENTITY.md, USER.md, MEMORY.md, TASKS.md, WORKSPACE.md
  agents/<name>/AGENT.md
  skills/<name>/SKILL.md          # Agent Skills layout
  sessions/<uuid>/state.json      # Serialized conversation (not long-term memory)
  memory/daily/                   # Short-term dated notes
  memory/weekly/, memory/monthly/ # Consolidation outputs
  logs/                           # Turn logs (input for consolidate)
  context/indexes/fragments.jsonl
  context/routing/, context/cache/
```

## Memory vs sessions

- **Sessions** hold chat state for one conversation (messages, tool results). Safe to delete without losing distilled memory if `MEMORY.md` / `memory/` are updated separately.
- **Memory** is cross-session knowledge. The **consolidate** tool summarizes logs + daily files into `MEMORY.md`.

## Priority

When instructions conflict: SOUL > per-agent AGENT body > IDENTITY > USER > MEMORY.
