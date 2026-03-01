# Requirements: REPL

## Role

The REPL is the client personality. It connects to the server over HTTP and runs an interactive loop: read a line from stdin, send it as a POST request, receive the streamed reply, **print each token as it arrives** (with stdout flushed after each token so output appears immediately), then wait for the next input.

## Behavior

- **No persistent connection**: Each user turn is one HTTP POST request. The request URL is from config (e.g. `http://127.0.0.1:8080/`).
- **Session**: On the first request, do not send `X-Session-Id`. Capture the session ID from the response header or first stream event. On every subsequent request, set the `X-Session-Id` header. If the server returns 401, clear the stored session ID and retry (or start a new session on next turn).
- **Loop**:
  1. Print a prompt (e.g. `> `).
  2. Read one line from stdin.
  3. POST the message (body: JSON `{"message":"..."}` or text/plain per config). Set `Accept: text/event-stream` (or NDJSON per config).
  4. Read the response stream (SSE or NDJSON); print each token delta to stdout **and flush stdout after each delta** so output appears immediately; capture session ID from header or first event if present.
  5. On `done` event/line, finish the turn. On `error` event/line, print error to stderr.
  6. Go to step 1.
- **Exit**: On EOF on stdin, or Ctrl+C (SIGINT), exit cleanly.

## Configuration

- **Server**: `AI_ASSISTANT_SERVER_ADDR` (host:port, default `127.0.0.1:8080`) or `AI_ASSISTANT_SERVER_URL` (full URL). Optional `AI_ASSISTANT_DEFAULT_REQUEST_TYPE` and `AI_ASSISTANT_DEFAULT_RESPONSE_TYPE`.

## Dependencies

- **Config**: REPL config (server URL/address, optional default request/response types).
- **Protocol**: Request body encoding, SSE/NDJSON stream parsing, session header name.

## Notes

- Empty lines can be ignored (no request sent).
- One HTTP request per user input; response is always streamed.
