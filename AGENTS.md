# Agent Guidelines for ai-assistant

TypeScript monorepo (Node 20+, `"type": "module"`). LLM stack: `@mariozechner/pi-ai`, `@mariozechner/pi-agent-core`; terminal UI: `@mariozechner/pi-tui`.

## Commands

```bash
npm install
npm run build          # core first, then server, repl, ask, consolidate
npm test               # vitest
npm run server
npm run repl
npm run ask -- "..."
npm run consolidate
```

## Layout

- `packages/core` — shared runtime (`SessionManager`, tools, workspace, indexer, protocol)
- `packages/server` — Hono app (`createApp()` in `src/server.ts`)
- `packages/repl`, `packages/ask`, `packages/consolidate` — thin CLIs

Workspace template lives in `packages/core/workspace-template/` and is copied into the user workspace on first run.

## Code style

- Prefer explicit types; avoid `any`
- Use `async`/`await` and native `fetch`
- Errors: throw `Error` with context or return HTTP errors from Hono handlers
- Match existing formatting; run `npx prettier --write` if the project adds Prettier later

## Testing

Vitest from repo root. Tests live next to sources: `*.test.ts`.

## Do not

- Reintroduce the removed Go tree (`internal/`, `cmd/`, `go.mod`)
- Edit the plan file in `.cursor/plans/` unless the user asks
