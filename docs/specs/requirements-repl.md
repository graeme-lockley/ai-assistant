# Requirements: REPL

## Implementation

- Package: `@ai-assistant/repl`.
- UI: `@mariozechner/pi-tui` (`ProcessTerminal`, `TUI`, `Text`, `Input`, `Container`).
- Transport: HTTP to server; prefers NDJSON stream for simpler line parsing.

## Session behavior

- Stores `X-Session-Id` after first response; sends on subsequent turns.
- `/agent <name>` clears session id so the next message opens a new session with that agent.

## Slash commands

| Command | Action |
|---------|--------|
| `/help` | Help text |
| `/exit` | Close session header + exit |
| `/agents` | `GET /agents` |
| `/models` | `GET /models` |
| `/model <id>` | `POST /model` |
| `/agent <name>` | Next message starts new session with agent |
