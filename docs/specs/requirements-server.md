# Requirements: Server

## Role

The server is the long-running personality that listens for HTTP. It maintains a **session store** (session ID → agent). Each session has one personal agent that maintains conversation history and uses the Deepseek API to stream replies.

## Behavior

- **Listen**: Bind to a configurable address (e.g. `:8080`) using `net/http.Server` (HTTP).
- **Single endpoint**: e.g. `POST /` for each chat turn.
- **Per request**:
  - **Session**: If `X-Session-Id` is present, look up the agent in the session store; if missing or invalid, respond with 401 Unauthorized. If no session ID, create a new session and agent, set `X-Session-Id` on the response.
  - **Request body**: Parse by `Content-Type` (application/json or text/plain); 415 for unsupported.
  - **Response**: Always streamed. Choose format from `Accept` (or config default): `text/event-stream` (SSE) or `application/json` (NDJSON). Call the agent’s `RespondStream`; for each chunk, encode as SSE or NDJSON and write to the response; flush. On completion send a done event/line; on error send an error event/line.
- **Shutdown**: On SIGINT/SIGTERM, stop accepting and shut down the HTTP server gracefully.

## Dependencies

- **Config**: Bind address, Deepseek API key, base URL, model, optional default response type (from environment).
- **Protocol**: ParseRequestBody, SSE/NDJSON encoders, header names.
- **Session**: Session store (create/lookup by ID).
- **Agent**: New(llmClient), RespondStream(ctx, userMessage, sendChunk).
- **LLM**: Single shared client created at startup; agents use it for streaming completion.

## Lifecycle

1. Load server config (env).
2. Create LLM client (Deepseek); fail if API key is missing.
3. Create session store with LLM client.
4. Start HTTP server with POST handler.
5. On signal, cancel context and shut down server.
