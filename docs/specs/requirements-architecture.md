# Requirements: Architecture

## Monorepo (`npm` workspaces)

| Package | Role |
|---------|------|
| `@ai-assistant/core` | Workspace bootstrap, agent loader, `SessionManager`, tools (`AgentTool`), protocol helpers, fragment indexer |
| `@ai-assistant/server` | Hono HTTP server; maps pi-agent-core events → SSE/NDJSON |
| `@ai-assistant/repl` | pi-tui terminal client |
| `@ai-assistant/ask` | Single-shot HTTP client |
| `@ai-assistant/consolidate` | Offline LLM pass to merge short-term memory into `MEMORY.md` |

## Data flow

1. Client `POST /` with `{ message, agent?, model? }` and optional `X-Session-Id`.
2. Server resolves or creates a session, loads or builds `pi-agent-core` `Agent` with filtered tools and composed system prompt.
3. `agent.prompt(message)` runs the tool loop; events stream to the client.
4. On completion, server appends a turn log and persists `sessions/<id>/state.json`.

## Dependencies

- Node.js 20+, ESM (`"type": "module"`).
- External: pi-ai, pi-agent-core, pi-tui (repl only), hono, commander, js-yaml, uuid, dotenv (CLIs).
