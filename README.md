# AI Assistant

HTTP server and REPL client for an AI assistant (Deepseek-backed). Streamed responses; session per client.

## Build

```bash
go build -o ai-assistant ./cmd/ai-assistant
```

## Run

**Server** (run in one terminal):

```bash
export DEEPSEEK_API_KEY=your-key
./ai-assistant server
```

Optional: `BIND_ADDR=:8080` (default `:8080`), `DEEPSEEK_BASE_URL`, `DEEPSEEK_MODEL`.

**REPL client** (run in another terminal; connects to the server):

```bash
./ai-assistant repl
```

Optional: `AI_ASSISTANT_SERVER_ADDR` or `AI_ASSISTANT_SERVER_URL` if the server is not on the default. The REPL supports input history (Up/Down to recall previous lines); history is stored under your config directory and limited to 1000 entries (`AI_ASSISTANT_REPL_HISTORY_FILE`, `AI_ASSISTANT_REPL_HISTORY_MAX`).

**Session logs** — When a client creates or closes a session, the **server** prints a line to its console (stderr), for example:

```
2025-03-01T12:00:00Z [session] created <uuid>
2025-03-01T12:05:00Z [session] closed <uuid> explicit
```

So you see these in the **terminal where the server is running**, not in the REPL terminal. Send a first message from the REPL (no session ID yet) to create a session; you should see `[session] created` in the server terminal.

## Docs

- **[docs/features/backlog.md](docs/features/backlog.md)** — Feature backlog and done list.
- **[docs/specs/](docs/specs/)** — Requirements and specs.
