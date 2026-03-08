# Requirements: Ask Command

## Role

The ask command is a single-shot client personality. It connects to the server over HTTP, sends one instruction, captures the response, and returns JSON with metadata. A new session is created for each request and closed after the response is received.

## Behavior

- **Single request**: One HTTP POST request with the user's instruction.
- **Session lifecycle**:
  1. Send POST request without `X-Session-Id` to create a new session.
  2. Capture session ID from response header (`X-Session-Id`).
  3. Process the streamed response.
  4. Send a POST request with `X-Session-Id` and `X-Session-Close: true` to close the session.
- **Output**: JSON with the following fields:
  - `entries`: Array of message entries, each with:
    - `type`: "thinking" or "output"
    - `content`: The text content
  - `model`: The model used (or empty if server default).
  - `session_id`: The session ID for this request.
  - `tokens`: Number of tokens received.
- **Error handling**: On error, output JSON with:
  - `error`: Error message.
  - `code`: Optional HTTP status code.
  - `details`: Optional additional context.

## Command-line Usage

```bash
./ai-assistant ask <instruction> [--model <model>]
```

- `instruction`: The instruction to send (required). Can be multiple words.
- `--model`: Optional model override. If not specified, uses server default.

## Configuration

- **Server**: `AI_ASSISTANT_SERVER_URL` (full URL, e.g. `http://127.0.0.1:8080`) or `AI_ASSISTANT_SERVER_ADDR` (host:port, default `127.0.0.1:8080`).
- **Model**: `AI_ASSISTANT_MODEL` (optional; overrides default model).
- **Request/Response types**: `AI_ASSISTANT_DEFAULT_REQUEST_TYPE` and `AI_ASSISTANT_DEFAULT_RESPONSE_TYPE` (same as REPL).

## Dependencies

- **Config**: Ask config (server URL, model, optional request/response types).
- **Protocol**: Request body encoding, SSE/NDJSON stream parsing, session header names.
- **Server**: Session creation and close endpoints.

## Notes

- The ask command is stateless between invocations - each call creates a fresh session.
- Output is always JSON, never streamed to stdout.
- Exit code is 0 on success, 1 on error.
