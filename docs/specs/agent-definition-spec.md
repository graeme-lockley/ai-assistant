# Agent definition spec (`AGENT.md`)

Each agent is a directory `agents/<dirName>/` with a required `AGENT.md`. Layout aligns with [Agent Skills](https://agentskills.io/specification) (YAML frontmatter + markdown body) and borrows ideas from [Cursor subagents](https://cursor.com/docs/subagents) (description-driven routing).

## Frontmatter (required)

| Field | Required | Notes |
|-------|----------|--------|
| `name` | yes | Lowercase, hyphens; must match parent directory name |
| `description` | yes | ≤1024 chars; when to use this agent |
| `tools` | no | List of built-in tool names; omit or empty = no tools |
| `skills` | no | Names of `skills/<name>/` folders this agent may use |
| `model` | no | Default model id if client does not pass `model` |

## Body

Free-form markdown: operating style, personality, when to use / not use, safety rules.

## Discovery

The server exposes `GET /agents` returning `{ name, description }[]` for routing UIs. Full body is loaded only when that agent is selected for a session.

## Built-in tool names

`web_search`, `web_get`, `exec_bash`, `read_file`, `read_dir`, `write_file`, `merge_file`.
