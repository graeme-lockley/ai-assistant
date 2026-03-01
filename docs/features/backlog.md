# Backlog

Planned features for the AI Assistant, in rough priority order.

## Backlog

| # | Feature | Spec | Status |
|---|---------|------|--------|
| 1 | **Tool collection** — Web search, web get, exec bash, read/write/merge file, read dir. Tools respect workspace constraints. | [tool-collection.md](backlog/tool-collection.md) | Backlog |

## Done

| Feature | Spec |
|---------|------|
| **REPL history** — Readline-style history in the REPL; Up/Down to navigate history, Left/Right within the line. History persisted across sessions. | [repl-history.md](done/repl-history.md) |
| **Streaming results** — All results streamed to the caller (no single-chunk responses). Request/response support multiple content types; session ID; HTTP with SSE/NDJSON; LLM client disables gzip and streams reasoning_content; REPL flushes stdout per token. | [streaming-results.md](done/streaming-results.md) |
| **Session console output** — Log to server console (with timestamp) when a session is created and when it is closed. Sessions have a defined lifecycle (created → active → closed). Explicit close via `X-Session-Close: true`. | [session-console-output.md](done/session-console-output.md) |

---

*Add new rows above; move items to Done or Backlog as needed.*
