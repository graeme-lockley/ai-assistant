# Requirements: Architecture

## Components

| Component | Path | Responsibility |
|-----------|------|----------------|
| Main | `cmd/ai-assistant/main.go` | Entrypoint; parses subcommand (`server` \| `repl`), loads config, invokes the appropriate personality. |
| Config | `internal/config` | Server bind address, Deepseek API key/base URL/model, optional default request/response types; REPL server URL/address. Loaded from environment variables. |
| Protocol | `internal/protocol` | HTTP: request/response structs, ParseRequestBody by Content-Type, SSE and NDJSON stream encoders, session header name. |
| Session | `internal/session` | In-memory session store (session ID → agent). Create and lookup by ID. |
| LLM | `internal/llm` | Deepseek client (OpenAI-compatible). HTTP client disables gzip for incremental streaming; streams both reasoning_content and content when present. Streams conversation history and assistant reply via callback. |
| Agent | `internal/agent` | Per-session personal agent. Holds message history; receives user message, calls LLM stream, forwards deltas to caller via RespondStream. |
| Server | `internal/server` | HTTP listen, POST handler; session lookup/create, request body parse, stream response (SSE or NDJSON). |
| REPL | `internal/repl` | HTTP client; loop: read line from stdin, POST request, read streamed response (SSE or NDJSON), print each token and flush stdout so output appears immediately; session ID in header. |

## Data flow

1. **REPL → Server**: For each line of user input, REPL sends an HTTP POST (body JSON or text/plain, optional `X-Session-Id`). Server parses body and looks up or creates session.
2. **Server → Session**: If no session ID, create new session and agent; else lookup agent by ID (401 if invalid).
3. **Server → Agent**: Call agent’s `RespondStream(ctx, message, sendChunk)`.
4. **Agent → LLM**: Agent appends user message to history and calls LLM’s `CompleteStream`; LLM streams content deltas (and reasoning_content when present; HTTP client has compression disabled for incremental delivery).
5. **Agent → Server**: For each delta, sendChunk is invoked; server encodes as SSE or NDJSON and writes to response, flushes.
6. **Server → REPL**: Streamed response (session event if new, token events, done or error). REPL parses stream, prints each token and flushes stdout, captures session ID.

## Diagram reference

See the plan document for the high-level sequence: HTTP request/streamed response, session store lookup, and agent streaming via LLM.
