# Requirements: Architecture

## Components

| Component | Path | Responsibility |
|-----------|------|----------------|
| Main | `cmd/ai-assistant/main.go` | Entrypoint; parses subcommand (`server` \| `repl`), loads config, invokes the appropriate personality. |
| Config | `internal/config` | Server bind address, Deepseek and/or Anthropic API keys, workspace root (default `~/.ai-assistant.workspace`), Tavily API key for web search, optional default request/response types; REPL server URL/address. Loaded from environment variables. |
| Protocol | `internal/protocol` | HTTP: request/response structs, ParseRequestBody by Content-Type, SSE and NDJSON stream encoders, session header name. |
| Workspace | `internal/workspace` | Resolve root, Ensure() from template, LoadBootstrap() (SOUL, AGENT, IDENTITY) for system prompt. |
| Session | `internal/session` | In-memory session store (session ID → agent). Create loads bootstrap and passes to agent; lookup by ID. |
| LLM | `internal/llm` | Multi-provider (Deepseek, Anthropic). Streaming with system prompt (date + bootstrap). HTTP client disables gzip for incremental streaming; streams reasoning_content and content when present. |
| Agent | `internal/agent` | Per-session personal agent. Holds bootstrap (system prompt), message history; receives user message, calls LLM stream (with tools when runner set), forwards deltas via RespondStream. |
| Tools | `internal/tools` | Runner: web_search (Tavily), web_get, exec_bash, read_file, read_dir, write_file, merge_file; paths relative to workspace root. |
| Server | `internal/server` | HTTP listen, POST handler; resolve workspace, Ensure(); session lookup/create (with root for bootstrap); request body parse, stream response (SSE or NDJSON). |
| REPL | `internal/repl` | HTTP client; loop: read line from stdin, POST request, read streamed response (SSE or NDJSON), print each token and flush stdout so output appears immediately; session ID in header. |

## Data flow

1. **REPL → Server**: For each line of user input, REPL sends an HTTP POST (body JSON or text/plain, optional `X-Session-Id`). Server parses body and looks up or creates session.
2. **Server → Session**: If no session ID, create new session (LoadBootstrap from workspace, New agent with bootstrap); else lookup agent by ID (401 if invalid).
3. **Server → Agent**: Call agent’s `RespondStream(ctx, message, sendChunk)`.
4. **Agent → LLM**: Agent appends user message to history and calls LLM with system prompt (bootstrap) and optional tool calls; LLM streams content deltas (and reasoning_content when present; HTTP client has compression disabled). Tool results are fed back until final reply.
5. **Agent → Server**: For each delta, sendChunk is invoked; server encodes as SSE or NDJSON and writes to response, flushes.
6. **Server → REPL**: Streamed response (session event if new, token events, done or error). REPL parses stream, prints each token and flushes stdout, captures session ID.

## Diagram reference

See the plan document for the high-level sequence: HTTP request/streamed response, session store lookup, and agent streaming via LLM.
