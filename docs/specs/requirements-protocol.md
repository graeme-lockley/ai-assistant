# Requirements: Protocol

## Transport

- **Layer**: HTTP (or HTTPS). The client sends one HTTP request per user turn; the server responds with a **streamed response body**. No WebSocket or other persistent duplex transport.
- **Endpoint**: Single endpoint for chat turns, e.g. `POST /` or `POST /chat`. Exact path is implementation-defined and documented.
- **Streaming only**: Every response is delivered as a stream of chunks (SSE or NDJSON). There is no API that returns the full result in a single response body. The server flushes after each token so the client receives data incrementally.

## Session

- **Header names**: Request and response use the header `X-Session-Id` for the session identifier.
- **Establishment**: On the **first** request (no `X-Session-Id` present), the server creates a session, creates an agent for that session, and communicates the session ID to the client in the response (e.g. in the `X-Session-Id` response header and/or in the first stream event).
- **Subsequent requests**: The client **must** send `X-Session-Id` with every request after the first. The server uses it to look up the existing agent/conversation.
- **Invalid or missing session**: If the client sends an unknown or expired session ID, or omits the session ID after one has been established, the server responds with **401 Unauthorized**. The client may then send a new request without `X-Session-Id` to start a new session.

## Request

- **Method**: POST.
- **Content types**: The server accepts at least two request content types:
  - **application/json** — Body is JSON: `{"message": "user input text"}`. `message` is a string; the user’s message.
  - **text/plain** — Body is the raw message bytes (UTF-8). No JSON wrapper.
- **Unsupported type**: If `Content-Type` is not supported, the server responds with **415 Unsupported Media Type**.

## Response (streamed)

- **Content types**: The server can respond with different stream formats according to client preference (`Accept` header or config default):
  - **text/event-stream** (SSE) — Event stream with typed events. See “SSE format” below.
  - **application/json** (NDJSON) — One JSON object per line. See “NDJSON format” below.
- **Negotiation**: Response format is chosen from `Accept` (e.g. `Accept: text/event-stream` or `Accept: application/json`). If the client does not send a relevant `Accept`, the server may use a configured default (e.g. `text/event-stream`).
- **End of stream**: A well-defined end signal (e.g. a `done` event or a final NDJSON line with `"type":"done"`) indicates completion; the server then closes the stream.

### SSE format (Content-Type: text/event-stream)

- Standard Server-Sent Events: each event has `event:` and optionally `data:` lines, separated by blank lines.
- Event types:
  - **session** — Sent once when a new session is created. `data:` is JSON: `{"session_id":"<id>"}`.
  - **token** — A content delta from the assistant. `data:` is JSON: `{"delta":"<text>"}`.
  - **done** — Stream completed successfully. No payload required.
  - **error** — An error occurred. `data:` is JSON: `{"error":"<description>"}`.
- End of stream: send a `done` event, then close the response body. For errors, send an `error` event (and optionally close).

### NDJSON format (Content-Type: application/json, streamed)

- One UTF-8 JSON object per line (newline-delimited JSON). Each line is a single object.
- Object types (discriminated by `type` field):
  - **session** — `{"type":"session","session_id":"<id>"}`.
  - **token** — `{"type":"token","delta":"<text>"}`.
  - **done** — `{"type":"done"}`.
  - **error** — `{"type":"error","error":"<description>"}`.
- End of stream: a line with `{"type":"done"}` or stream close.

## Implementation (HTTP/streaming)

- **internal/protocol** (or equivalent): Request/response structs for JSON body; helpers to parse request body by `Content-Type` (application/json, text/plain). Encoders for SSE events and NDJSON lines.
- **Server**: Uses session store, parses request by Content-Type, selects response format from Accept/config, and streams response using the chosen encoding until done or error.

## Legacy (TCP, deprecated)

The previous transport was TCP with length-prefixed JSON frames. That protocol is superseded by HTTP streaming for the main server and REPL. The frame-based encode/decode may remain in the codebase for reference or removal.
