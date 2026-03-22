# AI Assistant

TypeScript monorepo: HTTP gateway, REPL (pi-tui), one-shot `ask`, and overnight `consolidate` for memory. LLM and agent runtime use [pi-mono](https://github.com/badlogic/pi-mono) (`@mariozechner/pi-ai`, `@mariozechner/pi-agent-core`, `@mariozechner/pi-tui`).

## Prerequisites

- Node.js 20+
- At least one provider API key: `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, and/or `DEEPSEEK_API_KEY`
- Optional: `TAVILY_API_KEY` for `web_search`

## Install and build

```bash
npm install
npm run build
```

## Workspace

Default directory: `~/.ai-assistant.workspace` (override with `AI_ASSISTANT_WORKSPACE`). It is created from the template in `packages/core/workspace-template/` on first server start.

## Run

**Server** (terminal 1):

```bash
npm run server
# or: node packages/server/dist/index.js
```

**REPL** (terminal 2 — keep the server running in terminal 1):

```bash
npm run repl
```

The REPL is a pi-tui client: it must be able to reach the HTTP server (`AI_ASSISTANT_SERVER_URL` / `AI_ASSISTANT_SERVER_ADDR`, default `http://127.0.0.1:8080`). If the server is down, messages will fail or stall until the request times out.

**Ask** (single JSON line to stdout):

```bash
npm run ask -- "What is 2+2?"
```

**Consolidate** (nightly memory rollup; reads `memory/daily/` and `logs/`):

```bash
npm run consolidate
```

## Packages

| Package | Description |
|--------|-------------|
| `@ai-assistant/core` | Workspace, agents, tools, sessions, protocol, indexer |
| `@ai-assistant/server` | Hono HTTP API, streaming SSE/NDJSON |
| `@ai-assistant/repl` | pi-tui client |
| `@ai-assistant/ask` | HTTP client CLI |
| `@ai-assistant/consolidate` | LLM-based merge into `MEMORY.md` |

## Docs

- [docs/specs/](docs/specs/) — architecture and protocol
- [docs/features/backlog.md](docs/features/backlog.md) — roadmap placeholder

## Environment variables

| Variable | Purpose |
|----------|---------|
| `AI_ASSISTANT_BIND` | Server bind (default `:8080`) |
| `AI_ASSISTANT_WORKSPACE` | Workspace root |
| `AI_ASSISTANT_SERVER_URL` / `AI_ASSISTANT_SERVER_ADDR` | Client → server URL |
| `AI_ASSISTANT_DEFAULT_MODEL` | Default model id |
| `TAVILY_API_KEY` | Web search |
| `AI_ASSISTANT_BOOTSTRAP_RING2` | `false` to omit USER/MEMORY/TASKS from system prompt |

See `packages/core/src/config.ts` for the full list.
