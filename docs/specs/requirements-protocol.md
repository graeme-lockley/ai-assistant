# Requirements: Protocol

## Transport

HTTP only. One `POST` per user turn. Response body is always streamed (SSE or NDJSON). No WebSocket in baseline.

## Request

- **Headers**: `Content-Type: application/json` or `text/plain`; `Accept: text/event-stream` or `application/json` (NDJSON lines); optional `X-Session-Id`.
- **JSON body**: `{ "message": string, "model"?: string, "agent"?: string }` — `agent` applies only when creating a session.
- **Plain body**: raw UTF-8 message.

## SSE (`text/event-stream`)

Events: `session`, `agent`, `thinking`, `token`, `tool`, `done`, `error`. Payload JSON on `data:` lines (see legacy spec for `session`/`token` shapes).

## NDJSON (`application/json` stream)

One JSON object per line: `type` field discriminates (`session`, `agent`, `thinking`, `token`, `tool`, `done`, `error`).

## Response header

`X-Session-Id` set on every chat response.
