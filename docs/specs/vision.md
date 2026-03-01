# Vision

This project is a minimal, Openclaw-inspired AI assistant. It is a single binary with two personalities: a **server** that receives requests over **HTTP** and runs one personal agent per session, and a **REPL** client that connects to the server, sends user input, and receives the assistant’s reply as a **stream**.

## In scope

- **Single binary**: One executable; run mode is chosen by subcommand (`server` or `repl`).
- **Server personality**: Long-running HTTP server. Each session (identified by `X-Session-Id`) has one **personal agent** that holds conversation state and calls the Deepseek API to generate replies. All responses are **streamed** (SSE or NDJSON); there is no single-chunk response.
- **REPL personality**: Client that connects to the server via HTTP, reads a line from stdin, sends it as a POST request, receives the **streamed** reply (tokens printed as they arrive), and waits for the next input (loop until exit).
- **Deepseek only**: The server is configured to use the Deepseek API (OpenAI-compatible) as the sole LLM backend. The LLM client streams replies and disables response compression so tokens are delivered incrementally; when the model provides it (e.g. deepseek-reasoner), reasoning is streamed too.
- **Wire protocol**: HTTP (or HTTPS). One POST per user turn; streamed response body (Server-Sent Events or NDJSON). Session ID in request/response headers.

## Out of scope

- **Tools and plugins**: No tool execution, browser automation, or extensibility (in this vision; see backlog).
- **Multi-channel ingress**: Only the HTTP server and REPL client; no WhatsApp, Telegram, Discord, etc.
- **Authentication or TLS**: No auth or encryption in this minimal version (TLS can be added via reverse proxy).
- **WebSocket or raw TCP**: Transport is HTTP with streamed response body only.
- **Multi-provider / failover**: One LLM provider (Deepseek) only.

The design mirrors Openclaw’s idea of a central server and session-per-connection execution, but keeps the feature set to the above.
