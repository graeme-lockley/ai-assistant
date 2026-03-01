# Requirements: Deepseek

## Role

The server uses the Deepseek API as the only LLM backend. The API is OpenAI-compatible, so the implementation uses an OpenAI-style client with a custom base URL and API key.

## Configuration

- **API key**: Required. Set via `DEEPSEEK_API_KEY`. The server must not start without it.
- **Base URL**: Optional; default `https://api.deepseek.com`. Set via `DEEPSEEK_BASE_URL` if needed (e.g. for a proxy or alternate endpoint).
- **Model**: Optional; default `deepseek-chat`. Set via `DEEPSEEK_MODEL`. Other options include `deepseek-reasoner` for chain-of-thought.

## Client

- Use an OpenAI-compatible Go client (e.g. `github.com/sashabaranov/go-openai`) with:
  - Custom base URL set to the Deepseek API base URL.
  - API key in the client config (Bearer token).
  - **HTTP client**: Use a custom `HTTPClient` with `Transport: { DisableCompression: true }` so that streaming responses are not gzip-compressed; tokens are delivered incrementally as the API sends them instead of being buffered by decompression.
- **Streaming**: The agent uses `CreateChatCompletionStream` to stream the assistant reply. The client receives content deltas via `stream.Recv()` until EOF. For each chunk, both **reasoning_content** (e.g. from `deepseek-reasoner`) and **content** are forwarded to the caller so the user sees reasoning and final answer as they arrive. Each delta is encoded as SSE or NDJSON and flushed to the response.

## Behavior

- **Streaming**: Each agent turn streams the reply to the client. The LLM client calls `CreateChatCompletionStream`; the server forwards each content delta (and, when present, reasoning_content) to the HTTP response stream. The LLM HTTP client disables compression so tokens arrive incrementally.
- **Conversation history**: The agent passes the full history (user and assistant messages) on each call so the model has context.

## Errors

- Missing API key: server fails to start.
- API errors (rate limit, model error, etc.): reported back over the stream as an error event/line.
