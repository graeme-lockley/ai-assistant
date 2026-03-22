# Vision

Personal AI assistant with:

- **Gateway HTTP server** — one POST per turn, streamed responses (SSE or NDJSON), durable sessions any client can resume (REPL, `ask`, future mobile app).
- **Multiple agents** — each `agents/<name>/AGENT.md` defines instructions, allowed tools, and linked skills ([Agent Skills](https://agentskills.io/home)-style layout).
- **Workspace memory** — separate from sessions: daily notes, weekly/monthly rollups, root `MEMORY.md`, turn logs under `logs/`.
- **pi-mono stack** — `@mariozechner/pi-ai` for providers, `@mariozechner/pi-agent-core` for the tool loop, `@mariozechner/pi-tui` for the REPL.

Out of scope for the baseline: multi-tenant auth, hosted SaaS operations, non-HTTP transports.
