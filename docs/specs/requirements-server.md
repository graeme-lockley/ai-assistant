# Requirements: Server

## Stack

- **Runtime**: Node.js, `@hono/node-server`.
- **Entry**: `packages/server/src/index.ts` → `createApp()` in `server.ts`.

## Bind address

`AI_ASSISTANT_BIND` (default `:8080`). Parsed as `host:port` or `:port` (all interfaces).

## Routes

| Method | Path | Notes |
|--------|------|--------|
| `POST` | `/` | Chat turn; streamed body |
| `GET` | `/agents` | List `{ name, description }` |
| `GET` | `/models` | List allowed model ids for configured API keys |
| `GET` | `/sessions` | List session metadata |
| `GET` | `/sessions/:id` | Metadata + messages |
| `DELETE` | `/sessions/:id` | Remove session directory |
| `GET` | `/model` | Current model (requires `X-Session-Id`) |
| `POST` | `/model` | `{"model":"..."}` (requires session) |

## Session headers

- `X-Session-Id` — omit to create a session; required afterward.
- `X-Session-Close: true` with `X-Session-Id` — delete session, `204`.

## Streaming

Map pi-agent-core events to:

- `token` / `thinking` — text deltas
- `tool` — tool name at start
- `session`, `agent` — on new session
- `done`, `error` — terminal

## Startup

Ensure workspace directories exist; run fragment indexer when needed (`ensureIndex`).
