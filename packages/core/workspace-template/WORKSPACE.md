# WORKSPACE

Personal assistant workspace. Layout:

- **Ring 1**: `SOUL.md`, `IDENTITY.md`
- **Ring 2**: `USER.md`, `MEMORY.md`, `TASKS.md`
- **Agents**: `agents/<name>/AGENT.md` (per-agent instructions and tool/skill lists)
- **Skills**: `skills/<name>/SKILL.md` (Agent Skills format)
- **Sessions**: `sessions/<uuid>/state.json` (conversation state)
- **Memory**: `memory/daily/`, `memory/weekly/`, `memory/monthly/` + root `MEMORY.md`
- **Logs**: `logs/` (turn logs)
- **Context**: `context/indexes/fragments.jsonl`, `context/routing/`

Run `@ai-assistant/consolidate` nightly (see `skills/cron-scheduler/SKILL.md`) to roll short-term memory into `MEMORY.md`.
