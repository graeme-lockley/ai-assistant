# Requirements: Server

## Role

The server is the long-running personality that listens for HTTP. It maintains a **session store** (session ID → agent). Each session has one personal agent that maintains conversation history and uses the LLM (Deepseek and/or Anthropic) to stream replies. The workspace root is resolved at startup; new sessions are bootstrapped with workspace core (SOUL, AGENT, IDENTITY) in the system prompt.

## Behavior

- **Listen**: Bind to a configurable address (e.g. `:8080`) using `net/http.Server` (HTTP).
- **Single endpoint**: e.g. `POST /` for each chat turn.
- **Per request**:
  - **Session**: If `X-Session-Id` is present, look up the agent in the session store; if missing or invalid, respond with 401 Unauthorized. If no session ID, create a new session and agent, set `X-Session-Id` on the response.
  - **Request body**: Parse by `Content-Type` (application/json or text/plain); 415 for unsupported.
  - **Response**: Always streamed. Choose format from `Accept` (or config default): `text/event-stream` (SSE) or `application/json` (NDJSON). Call the agent’s `RespondStream`; for each chunk, encode as SSE or NDJSON and write to the response; flush. On completion send a done event/line; on error send an error event/line.
- **Shutdown**: On SIGINT/SIGTERM, stop accepting and shut down the HTTP server gracefully.

## Dependencies

- **Config**: Bind address, Deepseek and/or Anthropic API keys, workspace root (default `~/.ai-assistant.workspace`), Tavily API key for web search, optional default response type (from environment).
- **Workspace**: Resolve workspace root (expand `~`); Ensure() creates or repairs layout from embedded template; LoadBootstrap() reads SOUL.md, AGENT.md, IDENTITY.md for system prompt.
- **Protocol**: ParseRequestBody, SSE/NDJSON encoders, header names.
- **Session**: Session store (create/lookup by ID) with workspace root; on Create, load bootstrap and pass to agent.
- **Agent**: New(llmClient, runner, summarizer, bootstrap); RespondStream uses bootstrap as system prompt each turn; tool runner uses workspace root.
- **Tools**: Runner with workspace root and Tavily API key; web_search uses Tavily (POST); file tools and exec_bash use root.
- **LLM**: Multi-provider (Deepseek, Anthropic); streaming with optional system prompt (date + workspace bootstrap).

## Lifecycle

1. Load server config (env).
2. Resolve workspace root; run workspace.Ensure(root).
3. Create LLM client(s) (Deepseek and/or Anthropic); fail if both API keys missing.
4. Create tool runner (root, Tavily API key) and session store (LLM, runner, summarizer, root).
5. Start HTTP server with POST handler.
6. On signal, cancel context and shut down server.
