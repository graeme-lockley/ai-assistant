# Feature: Streaming results and multi-content-type support

**Status: Done**

## Summary

Change the server so that **all** result payloads are streamed back to the caller instead of being returned in a single response. There is no non-streaming mode: streaming is the only way to receive results. In addition, the protocol supports multiple request/response content types beyond `application/json` (e.g. `text/plain`, `text/event-stream`, or other negotiated types).

## Transport and connection (REPL/server use case)

**Decision: HTTP with streaming response** — Use standard HTTP (or HTTPS) with a **streamed response body** as the only transport. Each client "turn" is one HTTP request (e.g. POST with the command or message); the server responds with a single streamed response (e.g. SSE or chunked transfer) until end of stream. No WebSocket or other persistent duplex protocol in the initial design.

**Rationale (non-functional experience):**

- **Request/response fits the REPL model**: One user action (e.g. run command, send message) maps to one request and one streamed result. We do not need full-duplex or server-initiated pushes on the same channel for this.
- **Proxies and firewalls**: Plain HTTP streaming works with normal proxies, load balancers, and corporate firewalls. WebSockets often need special handling (e.g. upgrade, timeouts) and are sometimes blocked or inspected more strictly.
- **Reconnection and tooling**: With HTTP, each turn is a clear unit; retries are "send the request again." SSE adds optional built-in reconnection semantics. Standard HTTP is easy to debug (e.g. `curl`, browser devtools, any HTTP client).
- **Content negotiation**: Using `Content-Type` and `Accept` on normal HTTP keeps content-type handling simple and consistent with the rest of the web stack.
- **Operational simplicity**: One transport (HTTP), one set of timeouts and connection rules, no separate WebSocket path to maintain.

**When to reconsider WebSockets (out of scope for now):**

- **True bidirectional streaming**: Client must send data (e.g. interrupts, cancellation, or follow-up input) *while* the server is still streaming, without starting a new request.
- **Persistent session without new request per turn**: Many turns over one long-lived connection to avoid repeated connection setup (at the cost of more complex connection lifecycle and infra).

If those requirements appear later, we can add a WebSocket (or WebTransport) option alongside HTTP streaming; the streaming *content* format (e.g. SSE events or NDJSON lines) can be reused on either transport.

**Practical recommendation:** Prefer **SSE** (`text/event-stream`) for the default streaming response format: standard, tooling-friendly, and easy to extend with event types (e.g. token, tool_call, done). HTTP/2 can be used when available for connection multiplexing without changing the application protocol.

## Streaming behaviour

- **Streaming only**: The server never returns a complete result in one chunk. Every response is delivered as a stream of chunks (e.g. SSE, chunked transfer, or a custom framing).
- **Chunk semantics**: Chunks may represent partial assistant text, tool-call progress, structured events, or metadata; the exact shape is defined by the chosen content type and protocol.
- **End of stream**: A well-defined end signal (e.g. closing the stream or a dedicated end event) indicates completion so the client can finalise the response.

## Session ID

- **Establishment**: The server creates and owns the session. On the **first** request (no session ID present), the server establishes a session and communicates the **session ID** to the client (e.g. in a response header such as `X-Session-Id` or in the stream metadata).
- **Subsequent requests**: The client **must** send the session ID with **every** request after the first. The session ID is carried in a **request header** (e.g. `X-Session-Id`); the exact header name is part of the protocol documentation.
- **Behaviour**: The server uses the session ID to associate requests with the same logical session (e.g. conversation or REPL state). Requests that omit the session ID after one has been established, or that send an unknown or expired ID, are handled according to protocol rules (e.g. 401 Unauthorized or a new session created).

## Content types

- **Request**: The server accepts more than one `Content-Type` for incoming requests. At minimum:
  - `application/json` — structured JSON body (current behaviour).
  - At least one additional type (e.g. `text/plain` for simple text input, or another structured format) to be specified in the spec or config.
- **Response**: The server can respond with different content types according to client preference (e.g. `Accept` header or query/config):
  - `application/json` — stream of JSON values or NDJSON.
  - `text/event-stream` (SSE) — event stream with typed events and optional JSON payloads.
  - Optionally `text/plain` or other types for minimal/text-only streaming.
- **Negotiation**: Content type for both request and response is determined by explicit headers (and/or config) so that clients can choose the format they need.

## Integration

- **Protocol**: Define a small set of supported request and response content types; document the streaming format (framing, events, or chunk boundaries) for each. Document the **session ID** request and response header names and lifecycle.
- **Server**: All response paths use the same streaming mechanism; remove or refactor any code that returns a single full response body. Server creates and returns a session ID when none is provided; validates session ID on subsequent requests.
- **Config**: Optional defaults for request/response content type if not overridden by headers.

## Out of scope

- Non-streaming fallback or "single chunk" mode.
- Custom binary protocols; focus on text-based streaming (JSON, SSE, plain text).
- **WebSocket (or other persistent duplex transport)** in the initial design; see "Transport and connection" for when to reconsider.

## Acceptance criteria

- [x] **Transport**: All results are delivered over HTTP (or HTTPS) with a streamed response body; no WebSocket or other non-HTTP streaming transport is required.
- [x] **Session ID**: Server establishes a session and returns a session ID to the client (e.g. in a response header). The client includes this session ID in a request header on every subsequent request; the protocol documents the header names and behaviour when the ID is missing or invalid.
- [x] Every response from the server is streamed; there is no API that returns the full result in one chunk.
- [x] At least two request content types are supported (e.g. `application/json` and one other).
- [x] At least two response content types are supported (e.g. `application/json` stream and `text/event-stream`), selectable via headers or config.
- [x] Request and response content types are documented (including stream framing and end-of-stream semantics).
- [x] Clients can negotiate request/response format and receive a consistent, well-defined stream until end of stream.

## Implementation (done)

- **Transport**: HTTP POST per turn; response body is SSE (`text/event-stream`) or NDJSON (`application/json`), chosen via `Accept`. Server flushes after each token event so the client receives data incrementally.
- **Session**: Header `X-Session-Id`; server creates session on first request and returns ID in response header (and optionally in first stream event); client sends header on subsequent requests; 401 when invalid/expired.
- **LLM client**: HTTP client uses `Transport: { DisableCompression: true }` so the upstream API response is not gzip-compressed; tokens are delivered as the API sends them rather than buffered by decompression. Stream deltas include both `reasoning_content` (e.g. deepseek-reasoner) and `content`; both are forwarded so the user sees reasoning and final answer as they arrive.
- **REPL**: Uses a flushable stdout (e.g. `bufio.Writer` around `os.Stdout`); after each token delta from the stream, writes to the writer and flushes so output appears immediately on the terminal.
