# Specs

Specification documents for the AI Assistant project.

## Start here

- **[vision.md](vision.md)** — High-level vision, in-scope and out-of-scope features.

## Requirements

| Document | Description |
|----------|-------------|
| [requirements-architecture.md](requirements-architecture.md) | Components and data flow (main, config, protocol, llm, agent, server, repl). |
| [requirements-server.md](requirements-server.md) | Server personality: TCP listen, one agent per connection, lifecycle. |
| [requirements-repl.md](requirements-repl.md) | REPL personality: connect, read/send/receive/print loop, exit behavior. |
| [requirements-protocol.md](requirements-protocol.md) | Wire format: TCP, length-prefixed JSON frames, request/response/error shapes. |
| [requirements-deepseek.md](requirements-deepseek.md) | Deepseek integration: API key, base URL, model; no streaming in v1. |
| [requirements-config.md](requirements-config.md) | Configuration: env vars, defaults, server vs REPL. |
