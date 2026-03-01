# Feature: Session console output and lifecycle

**Status: Done**

## Summary

Introduce session lifecycle and server-console output so operators can see when sessions are created and closed. Sessions have an explicit lifecycle (created → active → closed) and the server logs these events to the console with a timestamp.

## Behaviour

- **New session**: When a new session is created, the server writes a message to the console (e.g. timestamp, session ID). This happens when a client establishes a session (e.g. first request or explicit session start).
- **Session closed**: When a session is closed (client disconnect, timeout, or explicit close), the server writes a message to the console (e.g. timestamp, session ID, reason/duration if useful).
- **Lifecycle**: Sessions are modelled with a clear lifecycle so that "created" and "closed" are well-defined events. Intermediate state (e.g. active, idle) may be tracked as needed to support logging and cleanup.

## Console output

- **On create**: Log line includes a **timestamp** (e.g. RFC3339), then session id: `2006-01-02T15:04:05Z07:00 [session] created <id>`.
- **On close**: Log line includes a **timestamp**, then session id and optional reason: `2006-01-02T15:04:05Z07:00 [session] closed <id> reason`. Avoid logging sensitive data.

## Integration

- **Server**: The component that manages sessions (or the connection handler) emits console output at lifecycle transitions. Session creation happens when a new session is allocated; session close happens on disconnect, timeout, or explicit close.
- **Lifecycle**: Sessions are created when needed (e.g. first request per connection or when a session endpoint is hit). They transition to closed when the client disconnects, a timeout fires, or the server receives an explicit close. No requirement for a separate "session service" in v1; lifecycle can be implemented in the existing connection/session handling code.
- **Explicit close**: Client sends `X-Session-Close: true` with `X-Session-Id` on a request (e.g. POST) to close the session; server responds 204 No Content.

## Out of scope

- Fancy formatting or external log aggregation (plain console lines are enough).
- Session persistence or resume across server restarts.
- Per-session log files (console only for v1).

## Acceptance criteria

- [x] When a new session is created, the server prints a clear message to the console (timestamp, session id).
- [x] When a session is closed, the server prints a clear message to the console (timestamp, session id, optional reason).
- [x] Sessions have a defined lifecycle: created → active → closed; creation and close are the events that trigger console output.
- [x] Console output does not include sensitive or PII data.
